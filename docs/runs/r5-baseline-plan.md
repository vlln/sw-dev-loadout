# Implementation plan — ledger v0.3.0 transfers

Target repo: `/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round5/baseline` (currently at tag `v0.2.0`, clean tree).

Everything below was checked against the real code in a throwaway copy under `/tmp`, not inferred from reading. Where a claim depends on actual behaviour, the observed output is quoted.

## 0. Verified starting state

```
$ GOFLAGS=-mod=mod GOPROXY=off go test ./...
?   	ledger/cmd/ledgerd	[no test files]
ok  	ledger/internal/api
ok  	ledger/internal/money
ok  	ledger/internal/store
```

`go test ./... -v | grep -c '^=== RUN'` → `17`. Go toolchain is `go1.26.4 darwin/arm64`, module `ledger`, stdlib only, no `go.sum`. Keep it that way: no new dependencies.

Facts about the existing code that drive the design (all measured):

| Fact | Evidence |
|---|---|
| `money.Parse` accepts `"0"`, `"-0.00"`, `"+12.50"`, and trailing/leading spaces (`" 3.00 "`, `"12.5 "`). | probe run |
| `money.Parse` **silently overflows**: `Parse("9223372036854775807")` → `-100`; `Parse("184467440737095516.20")` → `4` i.e. `0.04`. It computes `w*100 + f` with no range check. | probe run |
| `store.Post` consumes a `seq` from the global counter and appends; there is no multi-account primitive and no rollback. | `internal/store/store.go:49-60` |
| `store.Balance` is the **only** balance computation in the service; `api` renders it via `Cents.String()` in `handleAccounts`, `getAccount`, and `postEntry`. | `internal/store/store.go:63`, `internal/api/api.go:78,134,178` |
| `writeErr` maps an *internal* code to a status via a literal map, and writes a *different* string as the JSON `error` field. The two arguments are independent. There is no 409 entry. | `internal/api/api.go:34-45` |
| An unregistered top-level path is served by `http.ServeMux`'s default, **not** the JSON envelope: `POST /transfers` → `404`, `Content-Type: text/plain; charset=utf-8`, body `404 page not found`. | probe run |
| `/accounts/` handler already returns a JSON 404 for unknown sub-segments. | probe run |
| `Idempotency-Key` canonicalises normally; `r.Header.Get("Idempotency-Key")` returns the value. | probe run |
| `store.Store` is already mutex-guarded and `-race` clean under 50 concurrent `Post`s. | probe run with `-race` |

---

## 1. Design

### 1.1 Where the transfer lives, and why

The transfer must be atomic (R1) **and** must make its overdraft decision against the same balance the rest of the service reports (R6, R7). Those two constraints together decide the layering:

- Atomicity cannot be built in `internal/api` out of two `store.Post` calls. `Post` takes and releases `s.mu` per call, so another request can interleave between the debit and the credit, and if the second `Post` fails there is nothing to undo the first — exactly the "只扣款不入账" state R1 forbids. The only place where "both entries or neither" is expressible is inside the single critical section that owns `s.mu`, which is `internal/store`.
- The overdraft check has the same problem for a different reason. Checking the balance in `api` and then posting means read-then-write across two lock acquisitions: a concurrent transfer can slip in between and the floor is no longer enforced. R7 says "任何一次转账都必须对当前余额做透支检查", so the check must read the balance and commit the entries without releasing the lock.

So: **one new store method does validation, the overdraft check, both postings, and the balance readback, all under one lock hold.** `internal/api` does request parsing and HTTP status mapping only. No business rule is duplicated in `api`.

### 1.2 Balance has exactly one implementation (R6)

R6 is the requirement that most easily gets violated by accident. "整个服务里账户余额只能有一个计算口径" means I must not write a second summation loop inside the transfer path.

`Store.Balance` locks `s.mu`, so the transfer (which already holds it) cannot call it — that would deadlock, since `sync.Mutex` is not reentrant. The fix is to extract the loop into an unexported lock-free helper and have both callers use it:

```go
// balanceOf sums every entry on a. The caller must hold s.mu.
func balanceOf(a *Account) money.Cents
```

`Store.Balance` becomes `return balanceOf(a), nil`. The transfer calls `balanceOf` directly. There is then literally one summation loop in the binary, which is what R6 asks for. **Do not** add a `balanceLocked(id string)` variant that re-does the map lookup; `balanceOf` takes the `*Account` the caller already has.

### 1.3 Idempotency caches sequence numbers, not balances (R5 × R6 conflict)

This is the one genuine design conflict in FEATURE.md and it needs an explicit decision. R5 says a repeated key returns "首次的结果". R6 says `from_balance`/`to_balance` in a transfer response must equal what `GET /accounts/{id}` returns *immediately after*. Those two disagree whenever anything changes the balance between the original request and the replay.

Measured, with the naive "cache the whole response" implementation:

```
first=201  from_balance=-5.00
POST /accounts/cash/entries {"amount":"1.00"}      # unrelated single-account posting
replay=200 from_balance=-5.00   GET cash balance=-4.00   MATCH=false
```

R6 is violated. With "cache only `from`/`to`/`seq_from`/`seq_to` and recompute the balances on replay":

```
first=201  from_balance=-5.00
POST /accounts/cash/entries {"amount":"1.00"}
replay=200 from_balance=-4.00   GET cash balance=-4.00   MATCH=true
```

**Decision: R6 wins; recompute balances on replay.** Reasons: R6 is stated as an absolute invariant over the whole service ("只能有一个计算口径") while R5's "首次的结果" is naturally read as "the same postings, no new ones" — and that part is preserved exactly, because `seq_from`/`seq_to` are replayed verbatim and no entry is added. R6 also has a test that can catch a violation mechanically, and a stale cached balance is the kind of bug that silently lies to a client. Record this decision in the commit message.

Consequence to implement deliberately: the idempotency lookup happens **before** validation and before the overdraft check, so a replay of a transfer that was legal when first accepted still returns 200 even if the account has since dropped near the floor. Verified: first `100.00` transfer from a zero-balance `cash` → 201; replay of the same key → 200, not 409. Re-running the overdraft check on a replay would be wrong, since the entries already exist and re-charging them is not on the table.

The cache is keyed only by the header value. A reused key with a *different* body is not detected — see §5.4.

### 1.4 Positivity check lives in `money`

R2 needs "strictly > 0, and reject malformed". `money.Parse` already rejects malformed but happily returns `0` for `"0"` and negatives for `"-1.00"` (both measured). Adding the positivity test inside `money` rather than in `api` keeps amount semantics in the money package and gives the API a single call with a single error to branch on.

Also fix the silent overflow found above. It is not in FEATURE.md, but it directly undermines R2 and R7: `Parse("184467440737095516.20")` returns `0.04` today, so a request with that amount would pass a `> 0` check and post four cents — a validation bypass, not a cosmetic bug. The guard is one comparison before the multiply and is provably non-narrowing for every currently accepted input (verified: `"92233720368547758.07"`, the true `MaxInt64`, still parses; `"92233720368547758.08"` and `"9223372036854775807"` now error).

### 1.5 409 in the error envelope

`writeErr`'s status map has no 409 (confirmed at `internal/api/api.go:34-45`); an unmapped code falls through to `500`. Add one entry. Note the function's two string parameters: the first becomes the JSON `error` field, the second selects the status. R7 wants `error: "overdraft"` with status 409, so add `"overdraft": http.StatusConflict` to the map and call `writeErr(w, "overdraft", "overdraft", detail)`. Measured result: `409 {"error":"overdraft","detail":"..."}`.

### 1.6 Typed errors so `api` does not parse strings

The store returns distinguishable errors; `api` maps them to statuses with `errors.As`/`errors.Is`. R4 requires the detail to name *which* account is missing, so the error carries the id and the role rather than being a bare sentinel.

`AccountError` implements `Unwrap() error { return ErrNoAccount }` so existing `errors.Is(err, ErrNoAccount)` call sites keep working — `TestUnknownAccount` in `internal/store/store_test.go:26` relies on `errors.Is` against `ErrNoAccount`, and this keeps that contract available to any future caller.

### 1.7 Exact signatures of everything new

`internal/money/money.go` — add:

```go
// ErrRange is returned when a syntactically valid amount exceeds Cents.
var ErrRange = errors.New("amount out of range")

// ErrNotPositive is returned by ParsePositive for amounts that are zero or negative.
var ErrNotPositive = errors.New("amount must be positive")

// ParsePositive parses s and additionally requires a strictly positive value.
func ParsePositive(s string) (Cents, error)
```

and, inside `Parse`, immediately before `total := w*100 + f`:

```go
if w > (math.MaxInt64-f)/100 {
    return 0, ErrRange
}
```

(add `"math"` to the imports). `ErrRange` is a new error value, so `Parse`'s existing contract of "returns `ErrMalformed` for bad syntax" is unchanged for every input that parsed before.

`internal/store/store.go` — change:

```go
type Store struct {
    mu        sync.Mutex
    accounts  map[string]*Account
    nextSeq   int
    transfers map[string]transferRecord   // new
}
```

`New()` initialises `transfers: map[string]transferRecord{}`. Extract `balanceOf` as in §1.2.

`internal/store/transfer.go` — new file:

```go
// OverdraftFloor is the lowest balance a transfer may leave on the source account.
const OverdraftFloor = money.Cents(-10000) // -100.00

var (
    ErrSameAccount = errors.New("from and to must differ")
    ErrNotPositive = errors.New("amount must be positive")
)

// AccountError names the missing side of a transfer.
type AccountError struct {
    AccountID string
    Role      string // "from" or "to"
}
func (e *AccountError) Error() string  // "from account ghost does not exist"
func (e *AccountError) Unwrap() error  // ErrNoAccount

// OverdraftError reports the balance the rejected transfer would have produced.
type OverdraftError struct {
    AccountID string
    Resulting money.Cents
    Floor     money.Cents
}
func (e *OverdraftError) Error() string
// "transfer would leave account cash at -100.01, below the -100.00 floor"

// TransferResult is the outcome of one transfer.
type TransferResult struct {
    SeqFrom     int
    SeqTo       int
    FromBalance money.Cents
    ToBalance   money.Cents
    Replayed    bool // true when an idempotency key replayed an earlier transfer
}

// transferRecord remembers which postings an idempotency key produced.
// It deliberately stores no balances; see R6.
type transferRecord struct {
    From, To         string
    SeqFrom, SeqTo   int
}

// Transfer moves amount from one account to another as a single atomic step.
// An empty idemKey disables idempotency. amount must be strictly positive.
func (s *Store) Transfer(from, to string, amount money.Cents, memo, idemKey string) (TransferResult, error)
```

Body order inside the single `s.mu` hold — this ordering is load-bearing, implement it exactly:

1. if `idemKey != ""` and present in `s.transfers`: recompute both balances with `balanceOf`, return with `Replayed: true` and `nil` error. (Before validation — §1.3.)
2. `amount <= 0` → `ErrNotPositive`.
3. `from == to` → `ErrSameAccount`.
4. `from` not in `s.accounts` → `&AccountError{from, "from"}`. Check `from` before `to` so the "both missing" case has a deterministic answer.
5. `to` not in `s.accounts` → `&AccountError{to, "to"}`.
6. `after := balanceOf(src) - amount`; if `after < OverdraftFloor` → `&OverdraftError{from, after, OverdraftFloor}`.
7. Allocate `seqFrom` then `seqTo` from `s.nextSeq` (two increments), append `Entry{seqFrom, -amount, memo}` to `src` and `Entry{seqTo, +amount, memo}` to `dst`.
8. Build the result from `balanceOf(src)` / `balanceOf(dst)`, store a `transferRecord` under `idemKey` if non-empty, return.

Every rejection happens before step 7, so a failed transfer appends nothing **and** burns no sequence number. Verified: after a rejected overdraft and a rejected 404, the next successful transfer gets seqs `1,2`.

`internal/api/api.go` — add to the `writeErr` status map: `"overdraft": http.StatusConflict`. Register in `Routes()`:

```go
mux.HandleFunc("/transfers", s.handleTransfers)
mux.HandleFunc("/transfers/", func(w http.ResponseWriter, r *http.Request) {
    writeErr(w, "not_found", "not_found", "unknown path "+r.URL.Path)
})
```

The second registration exists because without it `POST /transfers/x` escapes to `ServeMux`'s plain-text 404 (measured), which breaks the "fixed JSON envelope" promise in `README.md:23`. With it: `404 {"error":"not_found","detail":"unknown path /transfers/x"}`, `Content-Type: application/json`.

New types and handler:

```go
type transferReq struct {
    From   string `json:"from"`
    To     string `json:"to"`
    Amount string `json:"amount"`
    Memo   string `json:"memo"`
}

type transferResp struct {
    SeqFrom     int    `json:"seq_from"`
    SeqTo       int    `json:"seq_to"`
    FromBalance string `json:"from_balance"`
    ToBalance   string `json:"to_balance"`
}

func (s *Server) handleTransfers(w http.ResponseWriter, r *http.Request)
```

Handler steps: reject non-POST with the existing `not_allowed` code; decode the body (`bad_request`, detail `"body must be a JSON object"`, matching `postEntry`); `money.ParsePositive(req.Amount)` with the detail chosen by `errors.Is` over `money.ErrNotPositive` / `money.ErrRange` / default malformed; reject empty `from` or `to`; reject `from == to`; read `r.Header.Get("Idempotency-Key")`; call `s.st.Transfer`; map errors with `errors.As` for `*store.AccountError` → `not_found` and `*store.OverdraftError` → `overdraft`, `errors.Is` for `store.ErrNotPositive`/`store.ErrSameAccount` → `bad_request`, default → `server_error`; status `201`, or `200` when `res.Replayed`; marshal `transferResp` (a struct, not a map, so field order in the JSON is stable).

The `from == to` and positivity checks exist in both layers on purpose. `api` checks so the 400 detail can quote the request, and `store` checks so the invariant holds for any future non-HTTP caller. The store is the authority; the API's copy is an early exit, not the rule.

`cmd/ledgerd/main.go` — no change needed. `Routes()` is the only wiring point.

`README.md` — add `POST /transfers` to the endpoint table and mention the 409 `overdraft` code and the `Idempotency-Key` header. Not a requirement, but the README currently claims the error envelope has no 409, and leaving it stale is a documentation bug introduced by this change.

---

## 2. Units of work

Four units. The boundaries are **not** stylistic; each one is forced by a compile-or-test dependency named below.

### Unit 1 — `internal/money`: positivity and range

Changes: `ParsePositive`, `ErrNotPositive`, `ErrRange`, overflow guard in `Parse`.
Tests added: `TestParsePositive`, `TestParseRejectsOverflow`.

**What forces this boundary:** `store.Transfer` and `handleTransfers` both reference `money.ParsePositive` and `money.ErrRange`. If this unit is not finished first, neither of the later units compiles. It is also the only unit that touches an existing function body (`Parse`), so landing it alone makes "did I break the 3 existing money tests?" a one-package question rather than a whole-repo one.

**Independently verifiable:** `go test ./internal/money/` passes with the 3 existing plus 2 new tests, and `go test ./...` stays green because nothing else has changed yet.

### Unit 2 — `internal/store`: `balanceOf` extraction + `Transfer`

Changes: `balanceOf`, `Store.transfers` field, `New()` init, new `transfer.go` with `Transfer`, `TransferResult`, `transferRecord`, `AccountError`, `OverdraftError`, `ErrSameAccount`, `ErrNotPositive`, `OverdraftFloor`.
Tests added: the 7 store tests in §4.2.

**What forces this boundary:** the atomicity and overdraft-ordering argument from §1.1 is entirely a store-level property — it is about what happens inside one `s.mu` hold, and it is testable with zero HTTP involvement. Splitting `balanceOf` into its own unit would be worse, not better: the extraction has no observable effect on its own (it is a pure refactor whose only purpose is to let `Transfer` reuse the loop), so a standalone unit for it would have no test that could fail. Conversely, merging it into Unit 3 would mean the atomicity tests could only be written through HTTP, where a failure cannot distinguish "store is wrong" from "handler is wrong".

**Independently verifiable:** `go test ./internal/store/ -race` passes; `go test ./...` still green because `api` has not changed and the `Balance` signature is untouched.

### Unit 3 — `internal/api`: endpoint, 409 mapping, routing

Changes: `writeErr` map entry, two `Routes()` registrations, `transferReq`, `transferResp`, `handleTransfers`.
Tests added: the 21 API tests in §4.3.

**What forces this boundary:** every HTTP-observable requirement (status codes, the `error`/`detail` envelope, the 200-vs-201 distinction, the R6 response-vs-GET equality) can only be asserted once a route exists, and this unit cannot be written before Unit 2 because it calls `s.st.Transfer` and type-switches on `*store.OverdraftError`. It is the last unit that can fail a FEATURE.md requirement.

**Independently verifiable:** full `go test ./... -race`, 17 original + 30 new tests.

### Unit 4 — `README.md`

Changes: endpoint table row, 409 and `Idempotency-Key` notes.

**What forces this boundary:** nothing in it is compiled or tested, and it must describe the final shape of Units 1-3. Putting it anywhere earlier means writing documentation for an interface that has not been pinned down yet; merging it into Unit 3 mixes a prose-only change into the commit that carries the behavioural risk.

**Why not a different number.** Three alternatives were considered and rejected. One unit per FEATURE.md requirement (7) does not work: R1, R2, R4, R6 and R7 all land in the same `Transfer` body and the same handler, so five of the seven units would touch the same two functions and could not be landed or reverted separately. Two units (store + api) hides the `money` change — the overflow guard modifies an existing function covered by existing tests, and that risk deserves to be isolated where it can be bisected. Five or more, by splitting `balanceOf` out, creates a unit with no failing test of its own. Four is the count at which every unit has both a compile-level reason to exist and at least one test that can fail if it is wrong.

**Order:** 1 → 2 → 3 → 4, forced by the compile dependencies (`api` needs `store.Transfer` and `*store.OverdraftError`; `store` needs nothing from `money` that is new — but `api` needs `money.ParsePositive`, and doing `money` first keeps every intermediate state of the tree buildable and green).

---

## 3. Requirement-by-requirement

| Req | How it is satisfied | How it is proven |
|---|---|---|
| **R1** `POST /transfers`, 201, both-or-neither | Route registered in `Routes()`; `transferResp` emits exactly `seq_from`, `seq_to`, `from_balance`, `to_balance`; both appends happen in step 7 of §1.7 after every rejection, inside one `s.mu` hold. | `TestTransferHappyPath` (shape + 201 + values), `TestTransferAtomicOnFailure` (rejected transfer leaves 0 entries on both accounts), `TestTransferSeqNotConsumedOnFailure` (store-level: seqs still `1,2` after two rejections), `TestTransferConcurrentSameKey` with `-race`. |
| **R2** amount strictly > 0 | `money.ParsePositive` rejects `<= 0`; `Parse` rejects malformed and now out-of-range; store re-checks `amount <= 0`. | `TestTransferRejectsNonPositiveAmount` (table: `"0"`, `"0.00"`, `"-0.01"`, `"-5"`, `"1.234"`, `"abc"`, `""`, `"184467440737095516.20"` → all 400 + `bad_request`), `TestParsePositive`, `TestParseRejectsOverflow`. |
| **R3** no self-transfer | `api` checks `req.From == req.To` → 400; `store` checks `from == to` → `ErrSameAccount`. | `TestTransferRejectsSelf` (400, `bad_request`, non-empty detail), `TestStoreTransferSameAccount`. |
| **R4** 404 naming the missing account | `AccountError{AccountID, Role}`; `from` checked before `to`; `api` writes `ae.Error()` as the detail. | `TestTransferUnknownAccount` — asserts 404, `not_found`, and that the detail **contains the missing id** and not the existing one, separately for a bad `from` and a bad `to`. Measured details: `"from account ghost does not exist"`, `"to account ghost does not exist"`. |
| **R5** idempotency key → 200, no new entries | `transfers` map consulted at step 1; replay returns cached seqs with `Replayed: true`; `api` maps that to 200. | `TestTransferIdempotentReplay` (first 201, second 200, identical `seq_from`/`seq_to`, entry count unchanged at 1 per account, balance unchanged), `TestTransferDistinctKeysBothApply`, `TestTransferNoKeyNotDeduplicated`, `TestTransferConcurrentSameKey` (40 concurrent, exactly one 201 and 39 200s, exactly 1 entry — measured). |
| **R6** one balance definition | Single `balanceOf`; `Store.Balance` delegates to it; `Transfer` calls it; replay recomputes rather than caching (§1.3). | `TestTransferBalanceMatchesGet` (response fields vs `GET /accounts/{id}` for both accounts), `TestTransferReplayBalanceMatchesGetAfterInterleavedPosting` — the test that fails under the naive cache (measured `-5.00` vs `-4.00`) and passes under the chosen design. |
| **R7** overdraft floor, 409 | `OverdraftFloor = -10000`; check at step 6 against `balanceOf(src)`, i.e. the live balance, on every transfer with no fast path around it; `"overdraft"` added to the status map. | `TestTransferOverdraftRejected` (409, `error == "overdraft"`, detail contains `"-100.01"`), `TestTransferAtFloorAllowed` (exactly `-100.00` is accepted → 201), `TestTransferOverdraftFromPositiveBalance` (seeded `+50.00`, transfer `150.00` lands exactly on the floor → 201; `150.01` → 409 with `-250.01` in the detail — this one fails any implementation that checks `amount` against the floor instead of the resulting balance), `TestStoreOverdraftUsesLiveBalance`. |

---

## 4. Tests to add

30 new tests: 2 in `money`, 7 in `store`, 21 in the new API test file. For each, the verdict against unmodified `v0.2.0` code. "Fails to compile" is the honest answer for most of them — at `v0.2.0` there is no `/transfers` route, no `Transfer`, no `ParsePositive` — and a test that cannot compile is a failing test. Where a test *would* compile and run against `v0.2.0`, that is called out explicitly, because those are the ones that pin existing behaviour.

### 4.1 `internal/money/money_test.go`

| Test | Asserts | vs v0.2.0 |
|---|---|---|
| `TestParsePositive` | `"0.01"`, `"1"`, `"12.50"` accepted; `"0"`, `"0.00"`, `"-0.01"`, `"-1"`, `"abc"`, `""` rejected; the first group errors `nil`, the second non-`nil`. | **Does not compile** — `ParsePositive` does not exist. |
| `TestParseRejectsOverflow` | `Parse("9223372036854775807")`, `Parse("92233720368547758.08")`, `Parse("184467440737095516.20")` all return a non-nil error; `Parse("92233720368547758.07")` still succeeds at `MaxInt64`. | **Compiles and FAILS.** Measured at `v0.2.0`: those three return `-100`, `-9223372036854775808` and `4` with `err=<nil>`. This is the pre-existing silent-overflow bug; the test documents it and Unit 1 fixes it. |

The 3 existing money tests are untouched and keep passing (verified after the guard was added).

### 4.2 `internal/store/store_test.go` (append; do not edit existing tests)

| Test | Asserts | vs v0.2.0 |
|---|---|---|
| `TestStoreTransferMovesBoth` | `Transfer("cash","bank",1250,"rent","")` → `SeqFrom < SeqTo`; `Balance("cash") == -1250`, `Balance("bank") == 1250`; one entry each. | Does not compile. |
| `TestStoreTransferSameAccount` | `errors.Is(err, ErrSameAccount)`; entry count on the account stays 0. | Does not compile. |
| `TestStoreTransferNonPositive` | amounts `0` and `-1` → `errors.Is(err, ErrNotPositive)`; no entries. | Does not compile. |
| `TestStoreTransferMissingAccount` | bad `from` → `*AccountError` with `Role == "from"`, `AccountID == "ghost"`; bad `to` → `Role == "to"`; both satisfy `errors.Is(err, ErrNoAccount)` via `Unwrap`. | Does not compile. |
| `TestStoreOverdraftUsesLiveBalance` | from a `0` balance, `Transfer` of `10000` succeeds (lands on the floor) and of `10001` returns `*OverdraftError` with `Resulting == -10001`, `Floor == -10000`; after seeding `+5000`, a `15001` transfer is rejected with `Resulting == -10001`. | Does not compile. |
| `TestTransferSeqNotConsumedOnFailure` | after a rejected overdraft and a rejected missing-account call, the next successful transfer returns `SeqFrom == 1`, `SeqTo == 2`. | Does not compile. Measured on the prototype: `ok seqs=1,2`. |
| `TestStoreTransferIdempotent` | same key twice → second has `Replayed == true` and identical seqs; `Entries` length is 1 on each account; a different key posts new entries with `Replayed == false`. | Does not compile. |

### 4.3 `internal/api/` — new file `transfer_test.go`

It reuses `newTestServer` and `do` from `api_test.go` (same package, so they are in scope; `api_test.go` itself is not modified). It adds one helper, which must live in the new file:

```go
func doKey(t *testing.T, h http.Handler, method, path, body, key string) (*httptest.ResponseRecorder, map[string]any)
```

— identical to `do` but sets `Idempotency-Key` when `key != ""`. A separate helper is needed because `do` has no way to attach a header and `api_test.go` must not be touched.

| Test | Asserts | vs v0.2.0 |
|---|---|---|
| `TestTransferHappyPath` | 201; `seq_from == 1`, `seq_to == 2`, `from_balance == "-12.50"`, `to_balance == "12.50"`; all four keys present. | **Compiles and FAILS**, because `POST /transfers` at `v0.2.0` returns the mux default. Note the failure mode: the body is `404 page not found`, not JSON, so `do` calls `t.Fatalf("body is not JSON")`. That is still a fail, but if you want a cleaner signal, assert `w.Code` before touching the parsed body. |
| `TestTransferBalanceMatchesGet` | response `from_balance`/`to_balance` equal `GET /accounts/cash` and `GET /accounts/bank` `balance`. | Compiles and fails (404). |
| `TestTransferRejectsNonPositiveAmount` | table of 8 bad amounts (§3 R2) → 400 and `error == "bad_request"` and non-empty `detail`; afterwards both accounts still have 0 entries. | Compiles and fails (404, and the status is 404 not 400). |
| `TestTransferRejectsSelf` | `from == to == "cash"` → 400, `bad_request`, detail mentions `cash`. | Compiles and fails. |
| `TestTransferRejectsMalformedBody` | body `not json` → 400, `bad_request`. | Compiles and fails (404 at `v0.2.0`). |
| `TestTransferRejectsMissingParties` | `{"amount":"1.00"}` with no `from`/`to` → 400. | Compiles and fails. |
| `TestTransferUnknownAccount` | bad `from`: 404, `not_found`, `strings.Contains(detail,"ghost")` and `!strings.Contains(detail,"bank")`; symmetric for bad `to`. | Compiles and fails on the `error`/`detail` assertions — `v0.2.0` returns 404 but with a plain-text body, so it dies at the JSON parse. |
| `TestTransferIdempotentReplay` | first 201, second (same key, same body) 200; `seq_from`/`seq_to` equal across both; `GET /accounts/cash/entries` has exactly 1 entry; `GET /accounts/cash` balance equals the replayed `from_balance`. | Compiles and fails. |
| `TestTransferDistinctKeysBothApply` | keys `k1`, `k2`, same body → both 201, different seqs, `cash` balance is twice the amount. | Compiles and fails. |
| `TestTransferNoKeyNotDeduplicated` | two identical requests with **no** header → both 201, distinct seqs, 2 entries. | Compiles and fails. |
| `TestTransferReplayBalanceMatchesGetAfterInterleavedPosting` | transfer with key `k1`; then `POST /accounts/cash/entries {"amount":"1.00"}`; then replay `k1` → 200 **and** `from_balance` equals a fresh `GET /accounts/cash`. | Compiles and fails. **This is the R5/R6 discriminator.** It also fails against the naive cached-response design (measured `-5.00` vs `-4.00`), which is exactly why it is in the suite. |
| `TestTransferReplayAfterFloorPressure` | transfer `100.00` from zero with key `k1` (lands on the floor, 201); replay `k1` → 200, not 409. | Compiles and fails. Pins the "do not re-run validation on replay" decision from §1.3. |
| `TestTransferOverdraftRejected` | from zero, transfer `100.01` → 409, `error == "overdraft"`, detail contains `"-100.01"`; both accounts still have 0 entries. | Compiles and fails. |
| `TestTransferAtFloorAllowed` | from zero, transfer exactly `100.00` → 201, `from_balance == "-100.00"`. | Compiles and fails. |
| `TestTransferOverdraftFromPositiveBalance` | seed `cash` with `+50.00` via `POST /accounts/cash/entries`; `150.00` → 201 with `from_balance == "-100.00"`; then `150.01` → 409 with `-250.01` in the detail. | Compiles and fails. Catches an implementation that compares `amount` to the floor instead of the resulting balance. |
| `TestTransferOverdraftAfterDirectPosting` | `POST /accounts/cash/entries {"amount":"-500.00"}` (the single-account endpoint has no floor, by design — §5.1), then transfer `0.01` → 409 with `-500.01` in the detail. | Compiles and fails. Documents the known hole and pins the transfer side's behaviour over it. |
| `TestTransferMethodNotAllowed` | `GET /transfers` and `DELETE /transfers` → 405, `error == "method_not_allowed"`. | Compiles and fails — `v0.2.0` gives 404 plain text. |
| `TestTransferSubPathIsJSON404` | `POST /transfers/` and `POST /transfers/x` → 404, `Content-Type: application/json`, `error == "not_found"`, non-empty `detail`. | Compiles and fails — `v0.2.0` returns `text/plain` `404 page not found`, so the JSON parse fails. |
| `TestTransferErrorsCarryDetail` | table of one 400, one 404 and one 409 request → each has a non-empty `detail`. Mirrors `TestEveryErrorCarriesDetail`. | Compiles and fails. |
| `TestTransferContentTypeIsJSON` | success response has `Content-Type: application/json`. | Compiles and fails. |
| `TestTransferConcurrentSameKey` | 40 goroutines POST the same body with the same key against a `cash` seeded well above the floor; exactly one 201, 39 × 200; `GET /accounts/bank/entries` has exactly 1 entry. Run under `-race`. | Compiles and fails. Measured on the prototype: `201=1 200=39 bank entries=1`, `-race` clean. |

Total after Unit 3: 2 money + 7 store + 21 API tests = 30 new, 47 with the 17 existing. **No existing test case is edited** (`api_test.go` is not touched at all; `money_test.go` and `store_test.go` gain appended functions only), satisfying "现有 17 个用例一个都不能改".

---

## 5. Problems in FEATURE.md, and the decisions taken

### 5.1 R7 is bypassable through the pre-existing endpoint — scope decision

R7 says "任何一次转账都必须对当前余额做透支检查". It constrains transfers only. `POST /accounts/{id}/entries` has no floor and accepts arbitrary negative amounts (verified: `{"amount":"-500.00"}` → 201, balance `-500.00`). So the floor is not an invariant of the ledger; it is a precondition on the transfer endpoint. A client can drive any account arbitrarily negative through the old endpoint and R7 will not have been violated.

**Decision: implement R7 exactly as written, on transfers only, and do not add a floor to `postEntry`.** Adding one there would change `postEntry`'s behaviour, and `TestEntriesRoundTrip` (`internal/api/api_test.go:71`) posts `-1.25` to a zero account while `TestPostAndBalance` (`internal/store/store_test.go:8`) posts `-150` — both are currently within the floor so they would still pass, but changing an endpoint the requirements did not mention, in a way that can reject requests it used to accept, is out of scope. `TestTransferOverdraftAfterDirectPosting` exists to make the hole visible and deliberate rather than accidental. Flag it to the reviewer; if the floor is meant to be a ledger invariant, that is a follow-up requirement and it needs a decision about the existing endpoint's contract.

### 5.2 R5 vs R6 — resolved in favour of R6

See §1.3 for the measurement and the reasoning. Short form: cache the seqs, recompute the balances, return 200. R5's "no new entries" is honoured exactly; R5's "首次的结果" is honoured for the sequence numbers but *not* for the balance strings when something changed in between, because R6 forbids that. This is the single place where the plan knowingly reads a requirement non-literally.

### 5.3 R5 does not say when idempotency is checked relative to validation

Not stated in FEATURE.md. Chosen: cache lookup first, so a replay never re-validates and never re-runs the overdraft check (§1.3). The alternative — validate then look up — would 409 a replay of an already-committed transfer whenever the account has since moved closer to the floor, which contradicts "返回首次的结果". Pinned by `TestTransferReplayAfterFloorPressure`.

### 5.4 R5 says nothing about key reuse with a different body

A client could send key `k1` with `{"amount":"5.00"}` and then `k1` with `{"amount":"500.00"}`. As specified, the second returns the first transfer's result. Real payment APIs return 422 on a key/body mismatch, but that status and error code are not in FEATURE.md and inventing them would add an envelope code nobody asked for.

**Decision: do not detect mismatches.** The record stores no amount, so detection is not even possible without adding a field. Document the limitation in the `Transfer` doc comment. Note the measured consequence: a replay of a key with an *invalid* body still 400s, because validation of `req.Amount` happens in `api` before `Transfer` is called (measured: `badreplay=400`). That is a mild inconsistency with §5.3's "replays skip validation" — the API-layer validation is not skipped, only the store-layer checks are. Accepted: a malformed request deserves a 400 regardless, and the key's cached transfer is not disturbed.

### 5.5 The idempotency cache grows without bound

In-memory, one entry per distinct key, never evicted. For this service (already fully in-memory, no persistence, no TTL mechanism anywhere) that matches the existing posture, and no requirement mentions expiry. Flagged, not fixed.

### 5.6 R7's `-100.00` applies to every account uniformly

FEATURE.md gives one global number with no per-account override. Implemented as the package constant `OverdraftFloor`. If per-account limits are wanted later this becomes a field on `Account`; no requirement asks for it now.

### 5.7 R1's example output is positional, not normative about seq allocation

The example shows `seq_from: 3, seq_to: 4` — consecutive, debit first. Nothing says the debit must get the lower number, but the example implies it and it is the natural reading. Implemented as debit-then-credit, consecutive. `TestStoreTransferMovesBoth` asserts `SeqFrom < SeqTo`; `TestTransferHappyPath` asserts `1` and `2` on a fresh store.

### 5.8 Not in FEATURE.md but fixed anyway: the `Parse` overflow

Justified in §1.4: at `v0.2.0` an 18-digit amount can parse to a small positive value, which is a bypass of R2's positivity gate and of R7's floor arithmetic. Left unfixed, `TestTransferRejectsNonPositiveAmount`'s `"184467440737095516.20"` case cannot pass. This is the only change to a pre-existing function body, which is why it is isolated in Unit 1.

### 5.9 `Content-Type` has no charset

`writeJSON` sets exactly `application/json` and `TestContentTypeIsJSON` (`internal/api/api_test.go:143`) asserts that exact string. Reuse `writeJSON`; do not "improve" it to `application/json; charset=utf-8` — that breaks an existing test, which the requirements forbid.

---

## 6. Acceptance steps

Exact commands. Run from the repo root:

```bash
cd /Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round5/baseline
```

Baseline, before any change:

```bash
GOFLAGS=-mod=mod GOPROXY=off go test ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./... -v 2>&1 | grep -c '^=== RUN'    # expect 17
```

After each unit:

```bash
GOFLAGS=-mod=mod GOPROXY=off go build ./...
GOFLAGS=-mod=mod GOPROXY=off go vet ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./...
```

Per-unit focus runs:

```bash
# Unit 1
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/money/ -v -count=1
# Unit 2
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/store/ -v -count=1 -race
# Unit 3
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/api/ -v -count=1 -race
```

Final gate:

```bash
GOFLAGS=-mod=mod GOPROXY=off go build ./...
GOFLAGS=-mod=mod GOPROXY=off go vet ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./... -count=1
GOFLAGS=-mod=mod GOPROXY=off go test ./... -count=1 -race
gofmt -l .                                    # must print nothing
git diff --stat v0.2.0 -- internal/api/api_test.go internal/money/money_test.go internal/store/store_test.go
```

The last command is the "17 tests unchanged" check. `api_test.go` must show **no** diff at all (the new API tests go in `internal/api/transfer_test.go`). `money_test.go` and `store_test.go` will show added lines only — confirm with:

```bash
git diff v0.2.0 -- internal/money/money_test.go internal/store/store_test.go | grep '^-[^-]'
```

which must print nothing (no removed lines, so no existing case was altered).

Test-count check:

```bash
GOFLAGS=-mod=mod GOPROXY=off go test ./... -v 2>&1 | grep -c '^=== RUN'    # expect 47
```

Manual smoke test of the end-to-end requirements (optional; the automated suite covers all of it):

```bash
GOFLAGS=-mod=mod GOPROXY=off go run ./cmd/ledgerd -addr 127.0.0.1:8080 &
curl -sS -i -X POST localhost:8080/transfers \
  -d '{"from":"cash","to":"bank","amount":"12.50","memo":"rent"}'        # 201
curl -sS localhost:8080/accounts/cash                                     # balance -12.50
curl -sS -i -X POST localhost:8080/transfers -H 'Idempotency-Key: k1' \
  -d '{"from":"cash","to":"bank","amount":"1.00"}'                        # 201
curl -sS -i -X POST localhost:8080/transfers -H 'Idempotency-Key: k1' \
  -d '{"from":"cash","to":"bank","amount":"1.00"}'                        # 200, same seqs
curl -sS -i -X POST localhost:8080/transfers \
  -d '{"from":"cash","to":"bank","amount":"1000.00"}'                     # 409 overdraft
curl -sS -i -X POST localhost:8080/transfers \
  -d '{"from":"cash","to":"ghost","amount":"1.00"}'                       # 404, detail names ghost
kill %1
```

---

## 7. Notes for the implementer

- Do not change `Store.Balance`'s signature; `api` calls it in three places and the store tests call it directly.
- `sync.Mutex` is not reentrant. Inside `Transfer` you hold `s.mu`; calling `s.Balance`, `s.Post` or `s.Entries` from there deadlocks. Use `balanceOf` and touch `a.entries` directly.
- Keep `transferResp` a struct. `postEntry` uses `map[string]any`, which sorts keys alphabetically in the JSON output; a struct preserves the field order shown in R1's example. Either satisfies the tests, but the struct matches the documented shape.
- `writeErr`'s first argument is the JSON `error` value and the second selects the HTTP status. They are different strings for `method_not_allowed`/`not_allowed`. For the new case both are `"overdraft"`.
- The prototype that validated every claim in this plan is at `/tmp/r5scratch/proto` (a copy, `.git` removed). It is scratch, not a deliverable; the diff against the baseline is a useful cross-check but the implementer should write the code from this document.
