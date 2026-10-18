package textutil

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"äöü", 2, "äö..."},
		{"", 3, ""},
	}
	for _, tt := range tests {
		if got := Truncate(tt.in, tt.n); got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}

func TestWords(t *testing.T) {
	if got := Words("  one two\tthree\n"); got != 3 {
		t.Errorf("Words = %d, want 3", got)
	}
}
