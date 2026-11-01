package money

import "testing"

func TestParseRoundTrip(t *testing.T) {
	cases := map[string]Cents{
		"0":      0,
		"1":      100,
		"1.5":    150,
		"1.05":   105,
		"-2.50":  -250,
		" 3.00 ": 300,
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("Parse(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{"", "abc", "1.234", ".5", "1.2.3", "--1"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) should fail", in)
		}
	}
}

func TestString(t *testing.T) {
	cases := map[Cents]string{0: "0.00", 150: "1.50", -205: "-2.05", 7: "0.07"}
	for in, want := range cases {
		if got := in.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", int64(in), got, want)
		}
	}
}
