package ratelimit

import "testing"

func TestAllowUntilEmpty(t *testing.T) {
	b := New(3, 1)
	for i := range 3 {
		if !b.Allow() {
			t.Fatalf("call %d denied", i)
		}
	}
	if b.Allow() {
		t.Error("call 4 allowed, want denied")
	}
}

func TestTickRefillsUpToCapacity(t *testing.T) {
	b := New(2, 5)
	b.Allow()
	b.Tick()
	allowed := 0
	for b.Allow() {
		allowed++
	}
	if allowed != 2 {
		t.Errorf("allowed %d after tick, want 2", allowed)
	}
}
