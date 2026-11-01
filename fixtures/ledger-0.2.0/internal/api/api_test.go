package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ledger/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	st := store.New()
	st.CreateAccount("cash", "Cash")
	st.CreateAccount("bank", "Bank")
	return New(st).Routes()
}

func do(t *testing.T, h http.Handler, method, path, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var parsed map[string]any
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("%s %s: body is not JSON: %q", method, path, w.Body.String())
		}
	}
	return w, parsed
}

func TestListAccounts(t *testing.T) {
	h := newTestServer(t)
	w, body := do(t, h, "GET", "/accounts", "")
	if w.Code != 200 {
		t.Fatalf("code = %d", w.Code)
	}
	accts, ok := body["accounts"].([]any)
	if !ok || len(accts) != 2 {
		t.Fatalf("accounts = %#v", body["accounts"])
	}
	first := accts[0].(map[string]any)
	if first["id"] != "bank" || first["balance"] != "0.00" {
		t.Errorf("first account = %#v", first)
	}
}

func TestPostEntryUpdatesBalance(t *testing.T) {
	h := newTestServer(t)
	w, body := do(t, h, "POST", "/accounts/cash/entries", `{"amount":"12.50","memo":"deposit"}`)
	if w.Code != 201 {
		t.Fatalf("code = %d, body = %#v", w.Code, body)
	}
	if body["balance"] != "12.50" {
		t.Errorf("balance = %#v", body["balance"])
	}
	_, got := do(t, h, "GET", "/accounts/cash", "")
	if got["balance"] != "12.50" {
		t.Errorf("after read balance = %#v", got["balance"])
	}
}

func TestEntriesRoundTrip(t *testing.T) {
	h := newTestServer(t)
	do(t, h, "POST", "/accounts/cash/entries", `{"amount":"5","memo":"a"}`)
	do(t, h, "POST", "/accounts/cash/entries", `{"amount":"-1.25","memo":"b"}`)
	w, body := do(t, h, "GET", "/accounts/cash/entries", "")
	if w.Code != 200 {
		t.Fatalf("code = %d", w.Code)
	}
	es := body["entries"].([]any)
	if len(es) != 2 {
		t.Fatalf("entries = %#v", es)
	}
	if es[1].(map[string]any)["amount"] != "-1.25" {
		t.Errorf("second entry = %#v", es[1])
	}
}

func TestUnknownAccountIs404(t *testing.T) {
	h := newTestServer(t)
	for _, path := range []string{"/accounts/ghost", "/accounts/ghost/entries"} {
		w, body := do(t, h, "GET", path, "")
		if w.Code != 404 {
			t.Errorf("GET %s code = %d", path, w.Code)
		}
		if body["error"] != "not_found" {
			t.Errorf("GET %s error = %#v", path, body["error"])
		}
	}
}

func TestBadAmountIs400(t *testing.T) {
	h := newTestServer(t)
	w, body := do(t, h, "POST", "/accounts/cash/entries", `{"amount":"1.234"}`)
	if w.Code != 400 || body["error"] != "bad_request" {
		t.Errorf("code = %d, body = %#v", w.Code, body)
	}
}

func TestMalformedBodyIs400(t *testing.T) {
	h := newTestServer(t)
	w, body := do(t, h, "POST", "/accounts/cash/entries", `not json`)
	if w.Code != 400 || body["error"] != "bad_request" {
		t.Errorf("code = %d, body = %#v", w.Code, body)
	}
}

func TestWrongMethodIs405(t *testing.T) {
	h := newTestServer(t)
	w, body := do(t, h, "DELETE", "/accounts/cash", "")
	if w.Code != 405 || body["error"] != "method_not_allowed" {
		t.Errorf("code = %d, body = %#v", w.Code, body)
	}
}

func TestEveryErrorCarriesDetail(t *testing.T) {
	h := newTestServer(t)
	cases := []struct {
		method, path, body string
	}{
		{"GET", "/accounts/ghost", ""},
		{"POST", "/accounts/cash/entries", `{"amount":"x"}`},
		{"DELETE", "/accounts/cash", ""},
		{"GET", "/accounts/cash/bogus", ""},
	}
	for _, c := range cases {
		_, body := do(t, h, c.method, c.path, c.body)
		if d, _ := body["detail"].(string); d == "" {
			t.Errorf("%s %s has no detail: %#v", c.method, c.path, body)
		}
	}
}

func TestContentTypeIsJSON(t *testing.T) {
	h := newTestServer(t)
	w, _ := do(t, h, "GET", "/accounts", "")
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
}
