#!/usr/bin/env python3
"""Round-5 black-box acceptance prober.

Written BEFORE either arm started. Neither arm can see it.
Probes only the public HTTP surface, so it is blind to internal design.

Self-check: on the unmodified 0.2.0 fixture exactly three checks may pass
(OLD-a, OLD-b, OLD-c). Those three are deliberate guard pins for behaviour
that already exists. Every other check must fail on 0.2.0; if one passes
there, it is not evidence of new behaviour and must be rewritten.

Usage: probe.py <repo-dir> [port]
"""
import json
import subprocess
import sys
import time
import urllib.error
import urllib.request

REPO = sys.argv[1]
PORT = int(sys.argv[2]) if len(sys.argv) > 2 else 8731
BASE = f"http://127.0.0.1:{PORT}"

results = []


def check(cid, desc, ok, got=""):
    results.append((cid, desc, bool(ok), got))


def req(method, path, body=None, headers=None):
    data = None
    if body is not None:
        data = body if isinstance(body, bytes) else json.dumps(body).encode()
    r = urllib.request.Request(BASE + path, data=data, method=method)
    r.add_header("Content-Type", "application/json")
    for k, v in (headers or {}).items():
        r.add_header(k, v)
    try:
        with urllib.request.urlopen(r, timeout=5) as resp:
            raw = resp.read()
            code = resp.status
    except urllib.error.HTTPError as e:
        raw = e.read()
        code = e.code
    try:
        parsed = json.loads(raw) if raw else {}
    except Exception:
        parsed = {"__raw": raw.decode(errors="replace")}
    return code, parsed


def entry_count(acct):
    _, b = req("GET", f"/accounts/{acct}/entries")
    return len(b.get("entries", []))


def main():
    build = subprocess.run(
        ["go", "build", "-o", "/tmp/r5probe-bin", "./cmd/ledgerd"],
        cwd=REPO, capture_output=True, text=True,
        env={**__import__("os").environ, "GOFLAGS": "-mod=mod", "GOPROXY": "off"},
    )
    if build.returncode != 0:
        print("BUILD FAILED\n" + build.stderr)
        sys.exit(2)

    proc = subprocess.Popen(
        ["/tmp/r5probe-bin", "-addr", f"127.0.0.1:{PORT}"],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
    )
    try:
        for _ in range(50):
            try:
                req("GET", "/accounts")
                break
            except Exception:
                time.sleep(0.1)
        else:
            print("SERVER NEVER CAME UP")
            sys.exit(2)

        # --- R1 happy path
        code, b = req("POST", "/transfers",
                      {"from": "cash", "to": "bank", "amount": "12.50", "memo": "rent"})
        check("R1-a", "happy transfer returns 201", code == 201, f"code={code}")
        check("R1-b", "response has seq_from and seq_to",
              "seq_from" in b and "seq_to" in b, f"keys={sorted(b)}")
        check("R1-c", "from_balance = -12.50", b.get("from_balance") == "-12.50",
              f"got={b.get('from_balance')}")
        check("R1-d", "to_balance = 12.50", b.get("to_balance") == "12.50",
              f"got={b.get('to_balance')}")
        check("R1-e", "seq_from != seq_to", b.get("seq_from") != b.get("seq_to"),
              f"{b.get('seq_from')} vs {b.get('seq_to')}")

        # --- R6 one balance path
        _, a1 = req("GET", "/accounts/cash")
        _, a2 = req("GET", "/accounts/bank")
        check("R6-a", "GET cash agrees with from_balance",
              a1.get("balance") == b.get("from_balance"),
              f"GET={a1.get('balance')} resp={b.get('from_balance')}")
        check("R6-b", "GET bank agrees with to_balance",
              a2.get("balance") == b.get("to_balance"),
              f"GET={a2.get('balance')} resp={b.get('to_balance')}")
        _, lst = req("GET", "/accounts")
        byid = {x["id"]: x for x in lst.get("accounts", [])}
        check("R6-c", "list view agrees too AND reflects the transfer",
              byid.get("cash", {}).get("balance") == a1.get("balance") == "-12.50",
              f"list={byid.get('cash',{}).get('balance')} single={a1.get('balance')}")

        # --- R1 atomicity: both entries landed
        check("R1-f", "cash got exactly 1 entry", entry_count("cash") == 1,
              f"n={entry_count('cash')}")
        check("R1-g", "bank got exactly 1 entry", entry_count("bank") == 1,
              f"n={entry_count('bank')}")

        # --- R2 positive amounts only
        for amt in ["0", "0.00", "-1", "-0.01", "abc", "1.234", ""]:
            code, bb = req("POST", "/transfers",
                           {"from": "cash", "to": "bank", "amount": amt})
            check(f"R2[{amt or 'empty'}]", "non-positive/malformed amount is 400",
                  code == 400 and bb.get("error") == "bad_request",
                  f"code={code} error={bb.get('error')}")

        # --- R3 self transfer
        code, bb = req("POST", "/transfers",
                       {"from": "cash", "to": "cash", "amount": "1.00"})
        check("R3", "self transfer is 400",
              code == 400 and bb.get("error") == "bad_request",
              f"code={code} error={bb.get('error')}")

        # --- R4 missing accounts, detail names which one
        code, bb = req("POST", "/transfers",
                       {"from": "ghost", "to": "bank", "amount": "1.00"})
        check("R4-a", "unknown from is 404",
              code == 404 and bb.get("error") == "not_found",
              f"code={code} error={bb.get('error')}")
        check("R4-b", "detail names 'ghost'", "ghost" in str(bb.get("detail", "")),
              f"detail={bb.get('detail')}")
        code, bb = req("POST", "/transfers",
                       {"from": "cash", "to": "phantom", "amount": "1.00"})
        check("R4-c", "unknown to is 404",
              code == 404 and bb.get("error") == "not_found",
              f"code={code} error={bb.get('error')}")
        check("R4-d", "detail names 'phantom'", "phantom" in str(bb.get("detail", "")),
              f"detail={bb.get('detail')}")

        # nothing above should have posted anything
        check("R2R3R4-noop", "rejected transfers posted no entries",
              entry_count("cash") == 1 and entry_count("bank") == 1,
              f"cash={entry_count('cash')} bank={entry_count('bank')}")

        # --- R5 idempotency
        key = "probe-key-1"
        c1, r1 = req("POST", "/transfers",
                     {"from": "cash", "to": "bank", "amount": "2.00"},
                     {"Idempotency-Key": key})
        check("R5-a", "first keyed request is 201", c1 == 201, f"code={c1}")
        n_cash, n_bank = entry_count("cash"), entry_count("bank")
        c2, r2 = req("POST", "/transfers",
                     {"from": "cash", "to": "bank", "amount": "2.00"},
                     {"Idempotency-Key": key})
        check("R5-b", "replay is 200 not 201", c2 == 200, f"code={c2}")
        check("R5-c", "replay body equals a SUCCESSFUL first body",
              c1 == 201 and r2 == r1 and "seq_from" in r1,
              f"c1={c1} first={r1} replay={r2}")
        check("R5-d", "replay added no entries (after a real first post)",
              c1 == 201 and n_cash > 0 and n_bank > 0
              and entry_count("cash") == n_cash and entry_count("bank") == n_bank,
              f"c1={c1} cash {n_cash}->{entry_count('cash')} bank {n_bank}->{entry_count('bank')}")
        c3, _ = req("POST", "/transfers",
                    {"from": "cash", "to": "bank", "amount": "2.00"},
                    {"Idempotency-Key": "probe-key-2"})
        check("R5-e", "a different key is a fresh 201", c3 == 201, f"code={c3}")
        check("R5-f", "unkeyed requests are never deduped",
              req("POST", "/transfers", {"from": "cash", "to": "bank",
                                         "amount": "1.00"})[0] == 201
              and req("POST", "/transfers", {"from": "cash", "to": "bank",
                                             "amount": "1.00"})[0] == 201,
              "one of the two unkeyed posts was not 201")

        # --- R7 overdraft, limit -100.00
        _, cur = req("GET", "/accounts/cash")
        check("R7-setup", "cash is above the floor before the probe",
              cur.get("balance") == "-19.50", f"balance={cur.get('balance')}")
        # -19.50 - 80.00 = -99.50, still allowed
        code, bb = req("POST", "/transfers",
                       {"from": "cash", "to": "bank", "amount": "80.00"})
        check("R7-a", "transfer to exactly above the floor is allowed", code == 201,
              f"code={code} body={bb}")
        # -99.50 - 0.50 = -100.00, the floor itself must be allowed
        code, bb = req("POST", "/transfers",
                       {"from": "cash", "to": "bank", "amount": "0.50"})
        check("R7-b", "landing exactly on -100.00 is allowed", code == 201,
              f"code={code} body={bb}")
        _, cur = req("GET", "/accounts/cash")
        check("R7-c", "cash is now exactly -100.00", cur.get("balance") == "-100.00",
              f"balance={cur.get('balance')}")
        n_cash, n_bank = entry_count("cash"), entry_count("bank")
        # one more cent must be refused
        code, bb = req("POST", "/transfers",
                       {"from": "cash", "to": "bank", "amount": "0.01"})
        check("R7-d", "crossing the floor is 409 (not 500)", code == 409,
              f"code={code}")
        check("R7-e", "error code is 'overdraft'", bb.get("error") == "overdraft",
              f"error={bb.get('error')}")
        check("R7-f", "detail states the resulting balance",
              "100.01" in str(bb.get("detail", "")), f"detail={bb.get('detail')}")
        check("R7-g", "a real 409 refusal posted no entries",
              code == 409 and n_cash > 0
              and entry_count("cash") == n_cash and entry_count("bank") == n_bank,
              f"code={code} cash {n_cash}->{entry_count('cash')} bank {n_bank}->{entry_count('bank')}")

        # --- regression: old endpoints still work
        code, bb = req("POST", "/accounts/bank/entries", {"amount": "1.00", "memo": "x"})
        check("OLD-a", "single-account posting still 201", code == 201, f"code={code}")
        code, bb = req("GET", "/accounts/ghost")
        check("OLD-b", "unknown account still 404", code == 404, f"code={code}")
        code, bb = req("DELETE", "/accounts/cash")
        check("OLD-c", "wrong method still 405", code == 405, f"code={code}")
        # envelope shape preserved on the new endpoint
        _, bb = req("POST", "/transfers", {"from": "cash", "to": "nope", "amount": "1"})
        check("OLD-d", "new endpoint uses the same error envelope",
              set(bb) <= {"error", "detail"} and bb.get("detail"),
              f"keys={sorted(bb)}")
    finally:
        proc.terminate()
        proc.wait(timeout=5)

    passed = sum(1 for _, _, ok, _ in results if ok)
    for cid, desc, ok, got in results:
        print(f"{'PASS' if ok else 'FAIL'}  {cid:16} {desc}" + (f"   [{got}]" if not ok else ""))
    print(f"\n{passed}/{len(results)} acceptance checks passed")
    sys.exit(0 if passed == len(results) else 1)


main()
