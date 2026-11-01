// Package api exposes the ledger over HTTP with a fixed JSON envelope.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ledger/internal/money"
	"ledger/internal/store"
)

// Server routes requests to a store.
type Server struct {
	st *store.Store
}

// New wires a server onto a store.
func New(st *store.Store) *Server { return &Server{st: st} }

// errorBody is the single shape every failure uses.
type errorBody struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code, errCode, detail string) {
	status := map[string]int{
		"not_found":    http.StatusNotFound,
		"bad_request":  http.StatusBadRequest,
		"not_allowed":  http.StatusMethodNotAllowed,
		"server_error": http.StatusInternalServerError,
	}[errCode]
	if status == 0 {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, errorBody{Error: code, Detail: detail})
}

// Routes returns the handler for every supported path.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/accounts", s.handleAccounts)
	mux.HandleFunc("/accounts/", s.handleAccount)
	return mux
}

type accountSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
}

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, "method_not_allowed", "not_allowed", "only GET is supported here")
		return
	}
	out := []accountSummary{}
	for _, id := range s.st.IDs() {
		name, err := s.st.Name(id)
		if err != nil {
			writeErr(w, "internal", "server_error", err.Error())
			return
		}
		bal, err := s.st.Balance(id)
		if err != nil {
			writeErr(w, "internal", "server_error", err.Error())
			return
		}
		out = append(out, accountSummary{ID: id, Name: name, Balance: bal.String()})
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": out})
}

type entryOut struct {
	Seq    int    `json:"seq"`
	Amount string `json:"amount"`
	Memo   string `json:"memo"`
}

type postEntryReq struct {
	Amount string `json:"amount"`
	Memo   string `json:"memo"`
}

// handleAccount serves /accounts/{id} and /accounts/{id}/entries.
func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/accounts/")
	id, tail, _ := strings.Cut(rest, "/")
	if id == "" {
		writeErr(w, "bad_request", "bad_request", "account id is required")
		return
	}
	switch tail {
	case "":
		if r.Method != http.MethodGet {
			writeErr(w, "method_not_allowed", "not_allowed", "only GET is supported here")
			return
		}
		s.getAccount(w, id)
	case "entries":
		switch r.Method {
		case http.MethodGet:
			s.listEntries(w, id)
		case http.MethodPost:
			s.postEntry(w, r, id)
		default:
			writeErr(w, "method_not_allowed", "not_allowed", "only GET and POST are supported here")
		}
	default:
		writeErr(w, "not_found", "not_found", "unknown path segment "+tail)
	}
}

func (s *Server) getAccount(w http.ResponseWriter, id string) {
	name, err := s.st.Name(id)
	if err != nil {
		writeErr(w, "not_found", "not_found", "account "+id+" does not exist")
		return
	}
	bal, err := s.st.Balance(id)
	if err != nil {
		writeErr(w, "not_found", "not_found", "account "+id+" does not exist")
		return
	}
	writeJSON(w, http.StatusOK, accountSummary{ID: id, Name: name, Balance: bal.String()})
}

func (s *Server) listEntries(w http.ResponseWriter, id string) {
	es, err := s.st.Entries(id)
	if err != nil {
		writeErr(w, "not_found", "not_found", "account "+id+" does not exist")
		return
	}
	out := []entryOut{}
	for _, e := range es {
		out = append(out, entryOut{Seq: e.Seq, Amount: e.Amount.String(), Memo: e.Memo})
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "entries": out})
}

func (s *Server) postEntry(w http.ResponseWriter, r *http.Request, id string) {
	var req postEntryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, "bad_request", "bad_request", "body must be a JSON object")
		return
	}
	amt, err := money.Parse(req.Amount)
	if err != nil {
		writeErr(w, "bad_request", "bad_request", "amount "+req.Amount+" is not a decimal value")
		return
	}
	seq, err := s.st.Post(id, amt, req.Memo)
	if err != nil {
		if errors.Is(err, store.ErrNoAccount) {
			writeErr(w, "not_found", "not_found", "account "+id+" does not exist")
			return
		}
		writeErr(w, "internal", "server_error", err.Error())
		return
	}
	bal, err := s.st.Balance(id)
	if err != nil {
		writeErr(w, "internal", "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"seq":     seq,
		"id":      id,
		"balance": bal.String(),
	})
}
