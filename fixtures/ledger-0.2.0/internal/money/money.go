// Package money handles amounts as integer cents so arithmetic never rounds.
package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Cents is a signed amount in minor currency units.
type Cents int64

// ErrMalformed is returned when a string is not a decimal amount.
var ErrMalformed = errors.New("malformed amount")

// Parse reads a decimal amount such as "-12.34" into Cents.
// Exactly zero, one or two fractional digits are accepted.
func Parse(s string) (Cents, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrMalformed
	}
	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}
	whole, frac, hasFrac := strings.Cut(s, ".")
	if !allDigits(whole) {
		return 0, ErrMalformed
	}
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, ErrMalformed
	}
	var f int64
	if hasFrac {
		switch len(frac) {
		case 1:
			frac += "0"
		case 2:
		default:
			return 0, ErrMalformed
		}
		if !allDigits(frac) {
			return 0, ErrMalformed
		}
		f, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, ErrMalformed
		}
	}
	total := w*100 + f
	if neg {
		total = -total
	}
	return Cents(total), nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// String renders the amount with two fractional digits.
func (c Cents) String() string {
	sign := ""
	v := int64(c)
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)
}
