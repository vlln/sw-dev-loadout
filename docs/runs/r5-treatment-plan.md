# Implementation plan — ledger v0.3.0 transfer endpoint

Target repo: `/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round5/treatment`, at tag `v0.2.0`.
Requirements source: `FEATURE.md` (R1–R7). Constraint from that file: `go test ./...` fully green, and
none of the 17 existing test cases may be edited.

Toolchain for every command in this plan (no network):

```
GOFLAGS=-mod=mod GOPROXY=off go test ./...
```

Go 1.26.4 darwin/arm64 confirmed present.

---

## 0. Measured baseline facts

Everything below was executed against the pristine `v0.2.0` tree (and, where noted, against a throwaway
copy under `/tmp`). These are observations, not predictions. The plan's design choices depend on them, so
an implementer who doubts one should re-run it rather than reason about it.

| # | Fact | How it was observed |
|---|---|---|
| F1 | 17 test cases pass; `cmd/ledgerd` has no test files. | `go test ./... -v`, counted 17 `=== RUN` lines |
| F2 | `go test -race ./...` passes. The race detector works on this host. | `go test -race ./...` |
| F3 | `POST /transfers` today returns **404 with `Content-Type: text/plain; charset=utf-8`** and body `404 page not found\n` — the net/http default, not the JSON envelope. | `httptest` probe against `New(st).Routes()` |
| F4 | Unknown top-level paths (`/foo`) already return text/plain 404. Unknown sub-segments under `/accounts/{id}/` return the JSON envelope. | same probe |
| F5 | `money.Parse` silently overflows. `Parse("184467440737095517")` returns **+84** cents with `err == nil`. `Parse("92233720368547758.08")` returns **-9223372036854775808** with `err == nil`. | `money.Parse` probe, values logged |
| F6 | `Parse` accepts `"+12.50"`→1250, `"12.5"`→1250, `"00012.50"`→1250, `"-0.00"`→0, `" 3.00 "`→300. It rejects `""`, `"."`-leading, `"12."`, `"1.234"`, `"1e2"`, `"99999999999999999999"`. | same probe |
| F7 | A `ServeMux` pattern `"POST /transfers"` makes other methods return **405 text/plain** with `Allow: POST`, bypassing the JSON envelope. `POST /transfers/` (trailing slash) returns text/plain 404. | separate `ServeMux` probe |
| F8 | `http.Header.Get("Idempotency-Key")` retrieves a header set as `idempotency-key` (canonicalisation works; no manual casing needed). | `httptest.NewRequest` probe |
| F9 | A test referring to `s.Transfer(...)` **does not compile** against v0.2.0: `vet: s.Transfer undefined (type *Store has no field or method Transfer)`. | `go vet ./internal/store/` with a probe test present |
| F10 | The range guard proposed in U1 keeps all 17 green, rejects the four overflow inputs, and still accepts the boundary value `"92233720368547758.07"`. | prototype in `/tmp/t5/proto`, full suite + targeted test |
| F11 | A store-level idempotency ledger under the existing mutex yields exactly **one** committed transfer and **one** destination entry when 64 goroutines race on the same key. | prototype, `go test -race -count=4`, logged `fresh=1 replays=63 dest_entries=1` |
| F12 | With the same key and a **different** request body, the prototype replays the first result. This is the literal R5 behaviour; see Problem P2. | same prototype, logged |
| F13 | `sync.Mutex` here is non-reentrant and `Store.Balance` locks internally, so `Transfer` cannot call the exported `Balance`. An unexported `balanceOf(a *Account)` helper that assumes the lock is held is required. | follows from `store.go:63-75`; prototype compiles and passes only with the helper |

The repo was left untouched: `git status --short --branch` reports `## main` with no modifications, and
`git describe --tags` still reports `v0.2.0`.

---

## 1. Design

### 1.1 The one decision that shapes everything

R1 (both entries or neither), R5 (a repeated key must add no entries), R6 (one and only one balance
caliber) and R7 (the overdraft check must run against the *current* balance) are four statements about
the same thing: the window between reading a balance and committing two entries must not be observable
or interruptible.

The existing store exposes only `Post` and `Balance`, each taking and releasing `s.mu` independently.
Any transfer composed at the API layer out of those primitives is a read-modify-write across three
separate critical sections. Between the `Balance` read and the second `Post`, another request can commit
entries, which breaks R7 (the check was against a stale balance) and R1 (a failure after the first
`Post` leaves a debit with no matching credit — exactly the "只扣款不入账" state the requirement forbids).

So the transfer is implemented as **one new store method that acquires `s.mu` once** and performs
validation, the overdraft check, both appends, the idempotency bookkeeping and the balance computation
inside that single critical section. The API layer does request shape validation and status mapping, and
never does ledger arithmetic.

This is also the honest answer to the state-guarantee question. The guarantee the new method offers is
**no effect on failure**: every rejection path returns before the first `append`, inside the lock, so no
externally observable state changed. That is a mechanism-backed claim, not an aspiration — there is no
compensation logic because none is needed, and none can be needed, as long as no early return is ever
added after the first `append`. The implementer must preserve that ordering; it is the whole proof.

### 1.2 Exact new surface

**`internal/money/money.go`** — one addition, one changed error value.

```go
// ErrRange is returned when a well-formed decimal amount does not fit in Cents.
var ErrRange = errors.New("amount out of range")
```

`Parse` keeps its signature `func Parse(s string) (Cents, error)`. Two changes inside it:

1. The existing `strconv.ParseInt(whole, 10, 64)` failure path returns `ErrRange` instead of
   `ErrMalformed` (the digits were already validated by `allDigits`, so the only remaining failure is
   magnitude).
2. Before computing `w*100 + f`, reject what would wrap:

```go
if w > math.MaxInt64/100 {
    return 0, ErrRange
}
total := w * 100
if f > math.MaxInt64-total {
    return 0, ErrRange
}
total += f
```

Add `"math"` to the imports. Both guards are needed: the first catches the `w*100` wrap, the second the
`+f` wrap at the very top of the range. Verified bounds (F10): `"92233720368547758.07"` is accepted and
`"92233720368547758.08"` is rejected.

**`internal/store/store.go`** — new exported surface.

```go
// MinBalance is the inclusive floor a transfer must leave the source account at or above.
const MinBalance money.Cents = -10_000

// ErrOverdraft reports that a transfer would push the source account below MinBalance.
var ErrOverdraft = errors.New("overdraft")

// AccountError names the account that was not found.
// It matches ErrNoAccount under errors.Is.
type AccountError struct{ ID string }

func (e *AccountError) Error() string
func (e *AccountError) Is(target error) bool

// OverdraftError carries the balance the source account would have reached.
// It matches ErrOverdraft under errors.Is.
type OverdraftError struct {
    Account   string
    Projected money.Cents
    Floor     money.Cents
}

func (e *OverdraftError) Error() string
func (e *OverdraftError) Is(target error) bool

// TransferResult reports both sequence numbers and both post-transfer balances.
type TransferResult struct {
    SeqFrom     int
    SeqTo       int
    FromBalance money.Cents
    ToBalance   money.Cents
}

// Transfer moves amount from one account to the other under a single lock
// acquisition. Both entries are appended or neither is.
//
// amount must be positive; the caller validates that. When idempotencyKey is
// non-empty and a previous Transfer committed under the same key, the stored
// result is returned with replay true and no entries are appended.
//
// On every error return no entry was appended and no balance changed.
// Errors: *AccountError (matches ErrNoAccount) when either id is unknown,
// checking from before to; *OverdraftError (matches ErrOverdraft) when the
// source would fall below MinBalance.
func (s *Store) Transfer(from, to string, amount money.Cents, memo, idempotencyKey string) (res TransferResult, replay bool, err error)
```

Plus one unexported helper and one new field.

```go
// balanceOf is the only place an account balance is computed.
// Callers must hold s.mu.
func balanceOf(a *Account) money.Cents
```

`Store` gains `idem map[string]TransferResult`, initialised in `New`. `Balance` is rewritten to
`return balanceOf(a), nil` — its body moves into the helper and its signature and behaviour are
unchanged. This is the mechanism for R6: after this change the summation loop exists exactly once in the
whole repo, and both `GET /accounts/{id}` and the transfer response read through it.

Order of operations inside `Transfer`, which is load-bearing:

1. idempotency-key lookup → return the stored result with `replay = true`
2. look up `from`, then `to` → `*AccountError`
3. `projected := balanceOf(src) - amount`; if `projected < MinBalance` → `*OverdraftError`
4. allocate `seqFrom` then `seqTo` from `s.nextSeq` (source gets the lower number, matching the R1 example)
5. append to source (`-amount`), append to destination (`+amount`), both with the same memo
6. compute both balances via `balanceOf`, store under the key if one was supplied, return

Note on step 3: the comparison is `projected < MinBalance`, so a transfer landing exactly on `-100.00`
is allowed. R7 says "不得低于 -100.00", which makes the floor inclusive.

Note on step 1: `Post`, `Balance`, `Entries` and `Name` keep returning the bare `ErrNoAccount` sentinel.
They are untouched. Only `Transfer` returns the richer `*AccountError`, because only R4 requires the
identity of the missing account in the response.

**`internal/api/api.go`** — new handler, new status mapping.

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

// handleTransfers serves POST /transfers.
func (s *Server) handleTransfers(w http.ResponseWriter, r *http.Request)
```

`Routes` gains `mux.HandleFunc("/transfers", s.handleTransfers)`.

`writeErr`'s status map gains one entry:

```go
"overdraft": http.StatusConflict,
```

Validation order in `handleTransfers`, which determines the status code on overlapping failures:

1. method is not POST → `writeErr(w, "method_not_allowed", "not_allowed", "only POST is supported here")`
2. body does not decode → `bad_request`, detail `"body must be a JSON object"`
3. `from` or `to` empty → `bad_request`
4. `money.Parse(req.Amount)` fails (either `ErrMalformed` or `ErrRange`) → `bad_request`
5. `amt <= 0` → `bad_request`, detail naming the amount
6. `req.From == req.To` → `bad_request`
7. call `st.Transfer`; map `*OverdraftError` → `overdraft`/409, `*AccountError` → `not_found`/404,
   anything else → `server_error`
8. success: 200 when `replay` is true, 201 otherwise

Steps 2–6 are stateless request-shape checks and step 7 is the only one that touches state. Keeping that
split is what lets the "no effect on failure" guarantee hold trivially for the whole 400 family.

Two deliberate non-changes in this file:

- `writeErr`'s odd two-string signature (the first argument becomes the JSON `error` field, the second
  selects the HTTP status) is left exactly as it is. `TestWrongMethodIs405` depends on the current
  `("method_not_allowed", "not_allowed")` split, so tidying the signature puts an existing test at risk
  for no behavioural gain. For overdraft, call it as `writeErr(w, "overdraft", "overdraft", detail)`.
- The route is registered as the plain pattern `"/transfers"`, not the Go 1.22 method pattern
  `"POST /transfers"`. Per F7 the method pattern produces a text/plain 405 with no `detail`, which
  contradicts the envelope that `README.md` documents as fixed. Dispatching the method inside the
  handler keeps every failure on `/transfers` inside the JSON envelope.

`POST /transfers/` with a trailing slash stays a text/plain 404 from the mux. That matches how the
service already treats unknown top-level paths (F4), so it is left alone rather than special-cased.

### 1.3 The R6 structural gate

R6 has two halves. "The transfer response must equal a following GET" is a behavioural claim and a test
proves it. "整个服务里账户余额只能有一个计算口径" is a claim about the shape of the source, and no
behavioural test can prove it — a second, accidentally-correct summation would pass every test while
leaving the next change one typo away from two calibers that disagree.

So that half gets a gate: a dependency-free Go test, `internal/store/caliber_gate_test.go`, that parses
every non-test `.go` file under `internal/` with `go/parser` and `go/ast`, finds every `*ast.AssignStmt`
whose `Tok` is `token.ADD_ASSIGN` and whose right-hand side selects a field named `Amount`, and asserts
there is exactly one, inside a function named `balanceOf`. A failure message must name the file, the
line, and the enclosing function.

The gate must prove it rejects. The same test file therefore contains a self-test that runs the detector
over two synthetic in-memory sources: one with two such assignments in two different functions, asserting
the detector reports both and the check fails; one with a single assignment inside `balanceOf`, asserting
it passes. Without the negative case a mistyped matcher would match nothing and stay green forever,
which is worse than having no gate.

This is the only gate being added, and that is a judgement: the repo has no CI, no `AGENTS.md`, no
generator, and no existing gate harness, so the adoption tier here is the lowest one. Building a word
budget gate, a generated-artifact `--check` mode, or a decision-record format checker on a repo of this
size and lifetime would be process for its own sake. R6 gets a gate because it is explicitly phrased as
a structural invariant; nothing else in FEATURE.md is.

### 1.4 Decision records

Two records, under `docs/decisions/implemented/`, each with a mandatory `Alternatives considered`
section:

- `docs/decisions/implemented/money/2026-11-01-parse-range-guard.md` — why `Parse` now rejects
  out-of-range decimals and why that is a prerequisite for R2 rather than a separate hardening task.
  Alternatives: leave `Parse` alone and range-check in the API (rejected: the overflowed value is already
  wrong and positive by then, so the API cannot distinguish it — see F5); widen `Cents` to `big.Int`
  (rejected: changes every signature and the README's stated int64-cents contract).
- `docs/decisions/implemented/store/2026-11-01-single-lock-transfer.md` — why the transfer lives in the
  store under one lock acquisition, and the resulting no-effect-on-failure guarantee. Alternatives:
  compose it in the API from `Balance` + two `Post` calls (rejected: breaks R1 and R7 as argued in §1.1);
  add a `Begin`/`Commit` transaction object to the store (rejected: a general transaction mechanism for a
  single two-account operation, with no second caller to justify it).

Two records and not four, because U1 and U2 each carry one genuine decision and U3 carries none — the
HTTP layer's choices (route pattern, validation order) are consequences of decisions recorded in the
store record, and splitting them would create records that only restate the diff.

---

## 2. Units, order, and what forced each boundary

**Three units, landing in this order.** Each is independently buildable, independently testable, and
independently revertable.

### U1 — `internal/money`: reject out-of-range amounts

Scope: `internal/money/money.go`, `internal/money/money_test.go` (new cases appended, existing ones
untouched), the parse decision record.

**What forced this boundary:** R2 requires that a non-positive or malformed amount be rejected with 400.
Against v0.2.0 that requirement has a branch with no correct answer. `Parse("184467440737095517")`
returns `+84` cents and no error (F5), so an API-layer check of the form `if amt <= 0` passes it and the
service commits a transfer of 84 cents in response to a request for 1.8e17. The requirement's words are
satisfied — the amount *is* strictly positive — while the service does something the requester never
asked for. The fix belongs in `Parse` because by the time the API sees the value the information needed
to detect the problem is gone.

It is a **separate unit** rather than part of U2/U3 because `money.Parse` is shared: the existing
`POST /accounts/{id}/entries` path calls it too (`api.go:156`). Changing a shared primitive has a
different blast radius from adding a new endpoint, and merging the two makes "did this break the
existing entry endpoint?" unanswerable from a single diff. It lands **first** because U3's positivity
check is only meaningful once this is in.

### U2 — `internal/store`: single-lock `Transfer`, one balance caliber, idempotency ledger

Scope: `internal/store/store.go`, `internal/store/store_test.go` (new cases appended),
`internal/store/caliber_gate_test.go` (new file), the store decision record.

**What forced this boundary:** this is the smallest set of changes that cannot be split without passing
through a broken intermediate state. Specifically:

- Extracting `balanceOf` and adding `Transfer` cannot be separated, because `Transfer` cannot call the
  exported `Balance` — the mutex is not reentrant and `Balance` locks internally (F13). The extraction
  is not an optional tidy-up; it is the only way `Transfer` can read a balance at all.
- The idempotency ledger cannot be deferred to U3. Putting the key cache in the API layer means checking
  the cache, then calling `Transfer`, then writing the cache, with no lock held across the three — two
  concurrent requests with the same key then both commit, violating R5's "不得新增任何分录". The
  prototype confirms the store-level version holds under 64 concurrent same-key requests (F11); the
  API-level version has no mechanism that could.
- The overdraft check cannot be deferred to U3 for the same reason as R7's wording: "对**当前**余额做
  透支检查". A check performed outside the lock is a check against a balance that may already be stale.

So the boundary is drawn where the lock is. Everything that must happen inside one critical section is
one unit. It lands **second** because it is what U3 calls, and **after** U1 so that the amount reaching
it is already known to be representable.

### U3 — `internal/api`: `POST /transfers` and the 409 envelope mapping

Scope: `internal/api/api.go`, `internal/api/api_test.go` (new cases appended).

**What forced this boundary:** a different surface with a different set of observers. U2's observers are
in-process Go callers; U3's are HTTP clients, and its contract is status codes, JSON field names and the
error envelope. It also touches `writeErr`'s status map, which every existing error path in the service
shares — a change that deserves to be visible on its own rather than buried in a store diff. It lands
**last** because it has nothing to call until U2 exists.

### Why three and not some other number

Not two: folding U1 into U2 or U3 would mix a fix to a shared primitive consumed by an existing endpoint
with a new feature, so a regression in `POST /accounts/{id}/entries` could not be attributed or reverted
separately.

Not four or more: the tempting extra cuts all produce intermediate states that are worse than no cut.
"Extract `balanceOf`" alone is a refactor with no consumer and no behaviour to verify. "Add `Transfer`
without idempotency" ships a method whose documented R5 contract is absent, and adding the cache
afterwards would mean re-touching the same critical section. "Route and validate, then map statuses"
ships an endpoint that answers with the wrong codes. Slicing is bounded by blast radius: the right
granularity is the one where a reviewer can judge each step on its own, and three steps reach that here.

---

## 3. Requirement-by-requirement: satisfaction and proof

| Req | How it is satisfied | How it is proven |
|---|---|---|
| **R1** endpoint, 201, both-or-neither | `POST /transfers` → `transferResp` with `seq_from`/`seq_to`/`from_balance`/`to_balance`; both appends happen inside one `s.mu` critical section with every rejection returning before the first append | `TestTransferHappyPath` (shape, 201, values); `TestTransferAtomicOnOverdraft` and `TestTransferAtomicOnMissingAccount` (entry counts on both accounts unchanged after a rejection); `TestStoreTransferAppendsBothOrNeither` |
| **R2** amount strictly positive | step 4 rejects parse failures (`ErrMalformed` or `ErrRange`), step 5 rejects `amt <= 0`; U1 makes `Parse` reject magnitudes that would wrap | `TestTransferRejectsNonPositiveAmount` over `"0"`, `"0.00"`, `"-0.01"`, `"-5"`; `TestTransferRejectsMalformedAmount` over `"abc"`, `"1.234"`, `""`, `"1e2"`; `TestTransferRejectsOutOfRangeAmount` over `"184467440737095517"`, `"92233720368547758.08"`; `TestParseRejectsOutOfRange` at the money layer |
| **R3** no self-transfer | step 6 compares `req.From == req.To` | `TestTransferRejectsSelfTransfer` (400, `error == "bad_request"`, non-empty detail) |
| **R4** 404 naming the missing account | `Transfer` returns `*AccountError{ID}`; the handler renders `"account " + e.ID + " does not exist"` | `TestTransferMissingAccountNamesIt` — asserts 404, `error == "not_found"`, and `strings.Contains(detail, "ghost")` for a missing `from` and separately for a missing `to` |
| **R5** idempotency key | `Transfer` consults and writes `s.idem` inside the same critical section; handler returns 200 when `replay` is true | `TestTransferIdempotentReplay` (second call → 200, byte-identical body, entry count unchanged); `TestStoreTransferIdempotencyIsAtomic` (64 goroutines, one key, `-race`, exactly one commit) |
| **R6** one balance caliber | `balanceOf` is the only summation; `Balance` and `Transfer` both read through it | `TestTransferBalancesMatchSubsequentGet` (response fields vs `GET /accounts/{id}` for both accounts, after two chained transfers); `TestBalanceCaliberIsSingular` (AST gate) plus its own self-test |
| **R7** overdraft floor, 409, projected balance in detail | `MinBalance = -10_000`; `projected < MinBalance` rejects; `OverdraftError.Projected` is rendered into the detail; `writeErr` maps `overdraft` → 409 | `TestTransferOverdraftIsConflict` (409, `error == "overdraft"`, detail contains `"-100.01"`); `TestTransferOverdraftBoundary` (exactly `-100.00` succeeds, one cent more fails); `TestStoreOverdraftCarriesProjectedBalance` |

---

## 4. Tests to add, with their measured baseline verdict

The baseline column is the result of **actually running** the test, or its probe equivalent, against
`v0.2.0` in a `/tmp` copy. Where a test's baseline failure is weak evidence, the column says so — a test
that fails only because a symbol or route is missing proves the feature is absent, not that the test
would catch a wrong implementation. For that reason each entry also names the **implementation mistake
it catches**, which is the property that actually matters once the feature exists.

### U1 — `internal/money/money_test.go`

**`TestParseRejectsOutOfRange`**
Asserts `errors.Is(err, ErrRange)` for `"184467440737095517"`, `"92233720368547758.08"`,
`"92233720368547759"`, `"99999999999999999999"`, `"-92233720368547758.08"`; and that
`"92233720368547758.07"`, `"-92233720368547758.07"`, `"0"`, `"12.50"` still parse to their exact values.
*Baseline: fails to compile* (`ErrRange` undefined), and with the sentinel stubbed in it fails on the
first three inputs, which return `nil` errors (F5). *Catches:* a guard placed only on `w*100` and not on
`+f` (the `.08` case), and a guard set one order of magnitude off (the `.07` case would start failing).

### U2 — `internal/store/store_test.go`

All five fail to compile against v0.2.0 — `s.Transfer undefined` (F9). That is a real but weak negative
control, so the mistake-catching property is stated for each.

**`TestStoreTransferAppendsBothOrNeither`**
Transfer 1250 between two fresh accounts; assert `SeqFrom < SeqTo`, source entry is `-1250`, destination
is `+1250`, both memos equal the supplied memo, and `Entries` lengths are 1 and 1.
*Catches:* reversed signs, a single entry, memo dropped, sequence numbers allocated in the wrong order.

**`TestStoreOverdraftCarriesProjectedBalance`**
From a zero balance, transfer `10_001`; assert `errors.Is(err, ErrOverdraft)`, `errors.As` yields
`*OverdraftError` with `Projected == -10_001` and `Floor == MinBalance`, and that both accounts still
have zero entries.
*Catches:* a projected value computed as `balance + amount` or reported as the pre-transfer balance; an
`OverdraftError` returned after the first append.

**`TestStoreTransferFloorIsInclusive`**
Transfer exactly `10_000` from zero: must succeed and leave `FromBalance == MinBalance`. Then one more
cent must fail.
*Catches:* `<=` where `<` belongs — the single most likely one-character error in this change.

**`TestStoreTransferMissingAccountIdentified`**
Unknown `from`, then unknown `to`, then both unknown; assert `errors.Is(err, ErrNoAccount)` and that
`errors.As` yields an `*AccountError` whose `ID` is `from` in the first and third cases and `to` in the
second.
*Catches:* a plain `ErrNoAccount` return that loses the identity R4 needs; a lookup order that names
`to` when both are missing.

**`TestStoreTransferIdempotencyIsAtomic`**
Fund an account, then launch 64 goroutines calling `Transfer` with one shared key; count results with
`replay == false`. Assert exactly 1, and that the destination has exactly 1 entry. Run under `-race`
with `-count=4`.
*Catches:* an idempotency check that is not inside the same critical section as the append. Measured
`fresh=1 replays=63 dest_entries=1` on the prototype (F11); an API-layer cache cannot produce that.
*Determinism note:* the assertion is on the terminal state after `wg.Wait()`, never on a sampled
intermediate, and the shared counter is guarded by its own mutex so the test itself is race-clean. No
`sleep` anywhere — `sync.WaitGroup` is the completion signal.

### U2 — `internal/store/caliber_gate_test.go` (new file)

**`TestBalanceCaliberIsSingular`**
Walks `internal/` non-test sources; asserts exactly one `+=`-into-`.Amount` assignment, inside
`balanceOf`.
*Baseline: fails* — against v0.2.0 the single occurrence is inside `Balance`, not `balanceOf`, so the
enclosing-function assertion rejects. *Catches:* a future second summation anywhere under `internal/`.

**`TestBalanceCaliberGateRejects`**
Runs the same detector over a synthetic source containing two such assignments and asserts the detector
reports both and the check fails with a message naming both functions; then over a compliant synthetic
source and asserts it passes.
*Baseline: passes* — and that is correct and intended. It tests the detector, not the repo, so it must
be green on any tree. Without it a matcher typo would make `TestBalanceCaliberIsSingular` vacuously
green forever.

### U3 — `internal/api/api_test.go`

Every one of these was run against v0.2.0 as a probe. **All of them fail, and all of them fail the same
way: HTTP 404 with a text/plain body, so the parsed JSON map is `nil`** (F3). That is honest but weak
evidence — it proves only that the route does not exist. The mistake-catching property is therefore the
load-bearing column here.

A local helper is needed because the existing `do` cannot set headers; add a sibling that takes a header
map and leave `do` untouched (editing it would count as changing an existing test file's behaviour).

**`TestTransferHappyPath`**
`POST /transfers` with `{"from":"cash","to":"bank","amount":"12.50","memo":"rent"}`; assert 201,
`from_balance == "-12.50"`, `to_balance == "12.50"`, `seq_from` and `seq_to` present with
`seq_from < seq_to`, and `Content-Type: application/json`.
*Catches:* wrong status, snake_case field names misspelled, balances swapped, cents rendered without two
decimals.

**`TestTransferRejectsNonPositiveAmount`** — `"0"`, `"0.00"`, `"-0.01"`, `"-5"` → 400 /
`error == "bad_request"` / non-empty detail.
*Catches:* `amt < 0` instead of `amt <= 0`, which lets `"0"` through — and note `"-0.00"` parses to 0
(F6), so it belongs in this set too.

**`TestTransferRejectsMalformedAmount`** — `"abc"`, `"1.234"`, `""`, `"1e2"`, absent field → 400.
*Catches:* a `Parse` error mapped to 500 instead of 400.

**`TestTransferRejectsOutOfRangeAmount`** — `"184467440737095517"`, `"92233720368547758.08"` → 400.
*Catches:* U1 silently reverted, or `ErrRange` mapped to `server_error`. This is the end-to-end proof
that R2 holds on the branch F5 exposed.

**`TestTransferRejectsSelfTransfer`** — `from == to == "cash"` → 400 / `bad_request`.
*Catches:* the self-check placed after the store call, which would commit two entries on one account.

**`TestTransferMissingAccountNamesIt`** — missing `from`, then missing `to` → 404 / `not_found` /
detail contains the missing id. Also asserts the *other* account gained no entries.
*Catches:* a generic "account does not exist" detail that satisfies `TestEveryErrorCarriesDetail` while
failing R4's actual requirement.

**`TestTransferOverdraftIsConflict`** — `"100.01"` from a zero-balance account → 409 /
`error == "overdraft"` / detail contains `"-100.01"`.
*Catches:* `overdraft` missing from the `writeErr` status map, which silently becomes 500 — the
`if status == 0 { status = 500 }` fallback at `api.go:41-43` makes that failure quiet. Also catches a
detail that omits the projected balance.

**`TestTransferOverdraftBoundary`** — `"100.00"` from zero succeeds with `from_balance == "-100.00"`;
a following `"0.01"` returns 409.
*Catches:* the inclusive/exclusive floor error, at the HTTP layer.

**`TestTransferIdempotentReplay`** — two identical requests with `Idempotency-Key: k1`; assert
201 then 200, byte-identical response bodies, and that `GET /accounts/cash/entries` reports exactly one
entry. Then a third request with **no** key must create new entries, proving the replay is keyed and not
a blanket dedupe.
*Catches:* a 201 on replay, a recomputed (and therefore different) body, a second pair of entries, and
an implementation that dedupes on request content instead of on the key.

**`TestTransferBalancesMatchSubsequentGet`** — two chained transfers, then compare `from_balance` and
`to_balance` from the last response against `GET /accounts/cash` and `GET /accounts/bank`.
*Catches:* R6 drift — a response balance computed from the request rather than read from the store.

**`TestTransfersRejectsNonPostWithEnvelope`** — `GET /transfers` and `DELETE /transfers` → 405,
`error == "method_not_allowed"`, non-empty `detail`, `Content-Type: application/json`.
*Catches:* the `"POST /transfers"` method-pattern route, which per F7 returns a text/plain 405 with no
detail and would break the envelope the README calls fixed. *Baseline note:* this one fails on v0.2.0
with 404/text-plain rather than 405, so it does not discriminate there — its value is entirely in
catching the route-registration shortcut.

### Tests deliberately not added

No test asserts a balance upper bound, and none asserts behaviour for a same-key-different-body request
(see P2 — the literal behaviour is implemented but its correctness is an open question for a human, and
pinning it in a test would make a disputed reading look settled).

---

## 5. Problems in FEATURE.md

### P1 — R2 is unsatisfiable on v0.2.0's parser (resolved inside this plan)

"amount 必须是严格大于 0 的十进制数" has a branch with no correct answer against the current code:
`Parse("184467440737095517")` yields `+84` cents with no error (F5). A positivity check at the API layer
returns 201 for that request and the service transfers 84 cents. The requirement's letter is met and the
behaviour is wrong.

**Decision:** fix `Parse` (U1) rather than range-check at the API layer, and map `ErrRange` to the same
400 `bad_request` that R2 specifies. The API layer cannot distinguish the overflowed value, because by
then it is an ordinary positive int64. This is in scope: without it R2 is not actually satisfiable.

Cost disclosed: U1 changes a shared parser used by the existing `POST /accounts/{id}/entries` endpoint.
Inputs that previously produced nonsense entries now produce 400. That is a behaviour change on an
existing endpoint, it is intentional, and it must be stated in the PR disclosure section. All 17 existing
tests still pass (F10).

### P2 — R5 is ambiguous for a reused key with a different body (needs a human decision)

R5 says a repeated key returns the first result, unconditionally. Read literally,
`Idempotency-Key: k` with `amount: "12.50"` followed by the same key with `amount: "9999.00"` returns
the first result, 200, and commits nothing — the caller is told their second, different request
succeeded. The prototype confirms this is what a literal implementation does (F12).

**Decision:** implement the literal reading, and do not add a test that blesses it. Writing a
request-fingerprint check that returns 409 on key conflict would be the safer API design, and it is what
I would recommend, but it contradicts the requirement as written and R5's wording is emphatic
("**同一个键重复请求时，返回首次的结果**"). Choosing the safer behaviour here would be an implementer
silently overriding a stated requirement on a money-handling path.

**What the implementer must do:** implement the literal behaviour, note it in the PR disclosure, and
raise the question for whoever owns FEATURE.md. Recommended resolution for a follow-up round: store a
hash of `(from, to, amount, memo)` alongside the cached result and return 409 `idempotency_conflict`
when a key is reused with different parameters.

### P3 — R7's floor is not a system invariant, only a transfer precondition

R7 constrains "任何一次转账". `POST /accounts/{id}/entries` has no floor check and keeps none under this
plan, so `POST /accounts/cash/entries` with `{"amount":"-500.00"}` still drives the balance to `-500.00`,
below the `-100.00` floor. A client that cannot overdraw via `/transfers` can overdraw via the older
endpoint.

**Decision:** implement R7 exactly as scoped — transfers only — and document the gap rather than close
it. Reasons: FEATURE.md scopes the check to transfers; extending it to `Post` changes the contract of a
shipped endpoint that no requirement asks to change; and `Post` is the only way to fund an account in
this service, so a floor there would need a separate "deposit vs withdrawal" concept that nothing asks
for. (For the record, adding the check to `Post` would not break any of the 17 existing tests — the
largest negative posting in them reaches `+350` — so this is a scope judgement, not a compatibility one.)

**What the implementer must do:** state in the PR that `MinBalance` is a transfer precondition, not an
account invariant, and that the single-entry endpoint can still cross it. If the requirement author
intended a global invariant, that is a new requirement and a new round.

### P4 — R5 does not say whether failures are cached (resolved by judgement)

Nothing says what a key does when the first request under it was rejected.

**Decision:** only committed transfers enter the idempotency ledger. A cached 404 or 409 would make a
later, legitimately retried request fail permanently for a reason that no longer holds — and R5's own
wording ("不得新增任何分录") presupposes that entries were created the first time. The `Transfer`
doc comment states this.

### P5 — the idempotency ledger grows without bound (accepted, documented)

`s.idem` has no eviction. A long-running process accumulates one `TransferResult` per distinct key
forever.

**Decision:** accept it for v0.3.0 and document it in the `Transfer` doc comment. A TTL needs a clock,
and injecting a clock to make it testable is more machinery than this requirement justifies. The store is
already wholly in memory and loses everything on restart, so the ceiling is one process lifetime. This is
a real limitation and belongs in the PR disclosure, not hidden.

### P6 — accumulated balances can still overflow (accepted, documented)

U1 bounds a single parsed amount, but nothing bounds the sum of many entries. Reaching the limit needs
roughly 9.2e16 cents of accumulated postings.

**Decision:** no check. Documented as a known limitation. Adding per-posting accumulation guards would
touch `Post`, `Balance` and `Transfer` to defend against a state no caller can plausibly reach.

### P7 — R3 and R4 overlap on a self-transfer of an unknown account (resolved by judgement)

`{"from":"ghost","to":"ghost"}` satisfies both R3 (400) and R4 (404). FEATURE.md does not order them.

**Decision:** 400. The self-transfer check is stateless request-shape validation and runs before any
store lookup, which keeps all validation in one phase ahead of all execution — the same split that makes
the no-effect-on-failure guarantee trivial for the whole 400 family. Documented in the handler.

### P8 — R4 does not say which account to name when both are missing (resolved by judgement)

**Decision:** name `from`. `Transfer` looks up `from` first and returns on the first miss, so the
behaviour is deterministic and testable rather than map-iteration dependent.

---

## 6. Acceptance steps

Run from the repo root,
`/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round5/plans/../treatment` — i.e.
`/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round5/treatment`.

**Before starting — confirm the baseline you are building on.**

```
git describe --tags
git status --short --branch
GOFLAGS=-mod=mod GOPROXY=off go test ./... -v 2>&1 | grep -c '^=== RUN'
```

Expect `v0.2.0`, a clean tree, and `17`.

**After U1.**

```
GOFLAGS=-mod=mod GOPROXY=off gofmt -l .
GOFLAGS=-mod=mod GOPROXY=off go vet ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/money/ -v -run 'TestParse|TestString'
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/api/ -run 'TestBadAmountIs400|TestPostEntryUpdatesBalance|TestEntriesRoundTrip'
```

`gofmt -l` must print nothing. The `internal/api` selection is the narrowest evidence that the shared
parser change did not regress the existing entry endpoint — that is the surface U1 touches beyond its own
package.

**After U2.**

```
GOFLAGS=-mod=mod GOPROXY=off gofmt -l .
GOFLAGS=-mod=mod GOPROXY=off go vet ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/store/ -v
GOFLAGS=-mod=mod GOPROXY=off go test -race -count=4 ./internal/store/ -run 'TestStoreTransferIdempotencyIsAtomic'
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/store/ -v -run 'TestBalanceCaliber'
```

The `-race -count=4` run is the only evidence that R5 holds concurrently; a single-threaded pass does not
establish it. Read the reported case count before treating a filtered run as evidence.

**After U3.**

```
GOFLAGS=-mod=mod GOPROXY=off gofmt -l .
GOFLAGS=-mod=mod GOPROXY=off go vet ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./internal/api/ -v
```

**Full acceptance.**

```
GOFLAGS=-mod=mod GOPROXY=off gofmt -l .
GOFLAGS=-mod=mod GOPROXY=off go vet ./...
GOFLAGS=-mod=mod GOPROXY=off go build ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./...
GOFLAGS=-mod=mod GOPROXY=off go test -race ./...
GOFLAGS=-mod=mod GOPROXY=off go test ./... -v 2>&1 | grep -c '^=== RUN'
git diff --stat v0.2.0 -- internal/money/money_test.go internal/store/store_test.go internal/api/api_test.go
```

Required outcomes: `gofmt -l` silent; `vet` and `build` clean; both `go test` invocations fully green;
the `=== RUN` count equal to 17 plus the number of cases added; and the final `git diff --stat` showing
only **insertions** in the three existing test files. A single deleted or modified line in those files
violates FEATURE.md's "现有 17 个用例一个都不能改" and must be reverted, not argued.

**Negative-control step — do not skip this.** The plan's baseline verdicts are measured, but for two
guards the only strong evidence is a deliberate break. On the finished tree, temporarily make each of
these single-character edits, confirm the named test fails, then revert:

1. `projected < MinBalance` → `projected <= MinBalance`.
   `TestStoreTransferFloorIsInclusive` and `TestTransferOverdraftBoundary` must fail.
2. Remove `"overdraft": http.StatusConflict` from `writeErr`'s map.
   `TestTransferOverdraftIsConflict` must fail with 500, not pass.

If either edit leaves the suite green, the test is not testing what it claims and must be fixed before
the change lands.

**Manual check of the R1 example, optional but cheap.**

```
GOFLAGS=-mod=mod GOPROXY=off go run ./cmd/ledgerd -addr 127.0.0.1:8099 &
curl -s -X POST 127.0.0.1:8099/accounts/cash/entries -d '{"amount":"1","memo":"a"}'
curl -s -X POST 127.0.0.1:8099/accounts/cash/entries -d '{"amount":"1","memo":"b"}'
curl -s -i -X POST 127.0.0.1:8099/transfers -d '{"from":"cash","to":"bank","amount":"12.50","memo":"rent"}'
curl -s 127.0.0.1:8099/accounts/cash
```

The transfer response should be 201 with `seq_from: 3` and `seq_to: 4`, matching FEATURE.md's example,
and the final `GET` must report the same `balance` as the transfer's `from_balance`. Kill the server
afterwards. This exercises the real binary rather than `httptest`, which is the only form in which the
routing and the envelope are observed by an actual consumer.

---

## 7. Review routing

**Risk: medium.** The change adds a money-moving endpoint, which is a sensitive surface, and it modifies
a parser shared with an already-shipped endpoint. Against that: the service is in-memory only, there is
no persistence, no authentication and no external consumer, so nothing here is irreversible and the blast
radius ends at process restart. Medium, not high, on those grounds — and medium, not low, because
P1 changes existing endpoint behaviour and P3 leaves a documented gap in the overdraft invariant.

Routing consequence: one competent human reviewer should decide, and three things must be on their desk
rather than buried in the diff — the P2 idempotency-semantics question (which genuinely needs an owner's
answer, not an implementer's), the P1 behaviour change to `POST /accounts/{id}/entries`, and the P3
scope gap. The remaining items in §5 are resolved by judgement and need no reply.

The PR disclosure section must list, each with its reason: the P1 behaviour change on an existing
endpoint; the P2 literal-reading decision; the P3 non-global floor; the P5 unbounded key ledger; and the
P6 unguarded balance accumulation. None of these should be discovered by a reader from the code.
