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
	if len(cfg.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(cfg.Checks))
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
}

func TestLoadDefaults(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)

	content := []byte(`defaults:
  tail: 40
  head: 10
  grep: "error"
  continue: true

checks:
  test:
    cmd: pytest -x
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
	if cfg.Defaults.Tail != 40 {
		t.Errorf("expected defaults.tail 40, got %d", cfg.Defaults.Tail)
	}
	if cfg.Defaults.Head != 10 {
		t.Errorf("expected defaults.head 10, got %d", cfg.Defaults.Head)
	}
	if cfg.Defaults.Grep != "error" {
		t.Errorf("expected defaults.grep 'error', got %q", cfg.Defaults.Grep)
	}
	if !cfg.Defaults.Continue {
		t.Error("expected defaults.continue true")
	}
}

func TestLoadDefaultsPartial(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)

	content := []byte(`defaults:
  tail: 20

checks:
  test:
    cmd: pytest -x
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
	if cfg.Defaults.Tail != 20 {
		t.Errorf("expected defaults.tail 20, got %d", cfg.Defaults.Tail)
	}
	if cfg.Defaults.Head != 0 {
		t.Errorf("expected defaults.head 0 (unset), got %d", cfg.Defaults.Head)
	}
	if cfg.Defaults.Grep != "" {
		t.Errorf("expected defaults.grep '' (unset), got %q", cfg.Defaults.Grep)
	}
	if cfg.Defaults.Continue {
		t.Error("expected defaults.continue false (unset)")
	}
}
