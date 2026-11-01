package store

import (
	"errors"
	"testing"
)

func TestPostAndBalance(t *testing.T) {
	s := New()
	s.CreateAccount("cash", "Cash")
	if _, err := s.Post("cash", 500, "open"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Post("cash", -150, "coffee"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Balance("cash")
	if err != nil {
		t.Fatal(err)
	}
	if got != 350 {
		t.Errorf("balance = %d, want 350", int64(got))
	}
}

func TestUnknownAccount(t *testing.T) {
	s := New()
	if _, err := s.Post("ghost", 1, "x"); !errors.Is(err, ErrNoAccount) {
		t.Errorf("Post: got %v", err)
	}
	if _, err := s.Balance("ghost"); !errors.Is(err, ErrNoAccount) {
		t.Errorf("Balance: got %v", err)
	}
	if _, err := s.Entries("ghost"); !errors.Is(err, ErrNoAccount) {
		t.Errorf("Entries: got %v", err)
	}
}

func TestSequenceIsGlobalAndIncreasing(t *testing.T) {
	s := New()
	s.CreateAccount("a", "A")
	s.CreateAccount("b", "B")
	s1, _ := s.Post("a", 1, "")
	s2, _ := s.Post("b", 1, "")
	s3, _ := s.Post("a", 1, "")
	if s1 >= s2 || s2 >= s3 {
		t.Errorf("sequence not increasing: %d %d %d", s1, s2, s3)
	}
}

func TestEntriesOrderedAndCopied(t *testing.T) {
	s := New()
	s.CreateAccount("a", "A")
	s.Post("a", 10, "one")
	s.Post("a", 20, "two")
	es, err := s.Entries("a")
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 2 || es[0].Memo != "one" || es[1].Memo != "two" {
		t.Fatalf("entries = %+v", es)
	}
	es[0].Memo = "mutated"
	again, _ := s.Entries("a")
	if again[0].Memo != "one" {
		t.Error("Entries returned an aliased slice")
	}
}

func TestIDsSorted(t *testing.T) {
	s := New()
	s.CreateAccount("z", "Z")
	s.CreateAccount("a", "A")
	ids := s.IDs()
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "z" {
		t.Errorf("IDs = %v", ids)
	}
}
