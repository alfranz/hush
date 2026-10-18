package config

import "testing"

func TestParse(t *testing.T) {
	m, err := Parse("# comment\nPORT = 8080\n\nHOST=localhost\n")
	if err != nil {
		t.Fatal(err)
	}
	if m["PORT"] != "8080" || m["HOST"] != "localhost" || len(m) != 2 {
		t.Errorf("Parse = %v", m)
	}
}

func TestParseError(t *testing.T) {
	if _, err := Parse("PORT\n"); err == nil {
		t.Error("want error for line without =")
	}
}

func TestInt(t *testing.T) {
	m := map[string]string{"PORT": "8080", "BAD": "x"}
	if v, _ := Int(m, "PORT", 0); v != 8080 {
		t.Errorf("Int(PORT) = %d", v)
	}
	if v, _ := Int(m, "MISSING", 42); v != 42 {
		t.Errorf("Int(MISSING) = %d", v)
	}
	if _, err := Int(m, "BAD", 0); err == nil {
		t.Error("want error for BAD")
	}
}
