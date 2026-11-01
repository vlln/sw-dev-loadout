// Package store keeps accounts and their entries in memory.
package store

import (
	"errors"
	"sort"
	"sync"

	"ledger/internal/money"
)

// ErrNoAccount is returned when an account id is unknown.
var ErrNoAccount = errors.New("no such account")

// Entry is one posting against one account.
type Entry struct {
	Seq    int         `json:"seq"`
	Amount money.Cents `json:"-"`
	Memo   string      `json:"memo"`
}

// Account is a named balance with its postings.
type Account struct {
	ID      string
	Name    string
	entries []Entry
}

// Store holds every account. It is safe for concurrent use.
type Store struct {
	mu       sync.Mutex
	accounts map[string]*Account
	nextSeq  int
}

// New returns an empty store.
func New() *Store {
	return &Store{accounts: map[string]*Account{}, nextSeq: 1}
}

// CreateAccount registers an account. An existing id is overwritten.
func (s *Store) CreateAccount(id, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[id] = &Account{ID: id, Name: name}
}

// Post appends an entry and returns its sequence number.
func (s *Store) Post(id string, amount money.Cents, memo string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok {
		return 0, ErrNoAccount
	}
	seq := s.nextSeq
	s.nextSeq++
	a.entries = append(a.entries, Entry{Seq: seq, Amount: amount, Memo: memo})
	return seq, nil
}

// Balance sums every entry on the account.
func (s *Store) Balance(id string) (money.Cents, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok {
		return 0, ErrNoAccount
	}
	var total money.Cents
	for _, e := range a.entries {
		total += e.Amount
	}
	return total, nil
}

// Entries returns the account's postings ordered by sequence.
func (s *Store) Entries(id string) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok {
		return nil, ErrNoAccount
	}
	out := make([]Entry, len(a.entries))
	copy(out, a.entries)
	sort.Slice(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out, nil
}

// IDs lists every account id in sorted order.
func (s *Store) IDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.accounts))
	for id := range s.accounts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Name returns the account's display name.
func (s *Store) Name(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok {
		return "", ErrNoAccount
	}
	return a.Name, nil
}
