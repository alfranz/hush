package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNoConfig(t *testing.T) {
	// Change to a temp dir with no config
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when no file exists")
	}
}

func TestLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)

	content := []byte(`checks:
  lint:
    cmd: ruff check .
  test:
    cmd: pytest -x
    tail: 40
    grep: "FAIL"
  types:
    cmd: mypy src/
    agent: true
`)
	os.WriteFile(filepath.Join(tmp, ".hush.yaml"), content, 0644)
	os.Chdir(tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if len(cfg.Checks) != 3 {
		t.Errorf("expected 3 checks, got %d", len(cfg.Checks))
	}
	lint := cfg.Checks["lint"]
	if lint.Cmd != "ruff check ." {
		t.Errorf("unexpected lint cmd: %q", lint.Cmd)
	}
	test := cfg.Checks["test"]
	if test.Tail != 40 {
		t.Errorf("expected tail 40, got %d", test.Tail)
	}
	if test.Grep != "FAIL" {
		t.Errorf("expected grep FAIL, got %q", test.Grep)
	}
	types := cfg.Checks["types"]
	if !types.Agent {
		t.Error("expected agent mode for types check")
	}
}
