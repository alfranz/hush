package filter

import "testing"

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello world", "hello world"},
		{"red text", "\x1b[31merror\x1b[0m", "error"},
		{"bold", "\x1b[1mbold\x1b[0m", "bold"},
		{"multiple codes", "\x1b[1;31mred bold\x1b[0m", "red bold"},
		{"mixed", "before \x1b[32mgreen\x1b[0m after", "before green after"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(StripANSI([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
