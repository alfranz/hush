package cache

import "testing"

func TestPutGet(t *testing.T) {
	c := New(2)
	c.Put("a", 1)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("Get(a) = %d, %v; want 1, true", v, ok)
	}
}

func TestEvictsLeastRecentlyUsed(t *testing.T) {
	c := New(2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a") // a is now newer than b
	c.Put("c", 3)

	if _, ok := c.Get("b"); ok {
		t.Errorf("b should be evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Errorf("a should still be cached")
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
}

func TestUpdateKeepsSize(t *testing.T) {
	c := New(1)
	c.Put("a", 1)
	c.Put("a", 2)
	if v, _ := c.Get("a"); v != 2 || c.Len() != 1 {
		t.Errorf("Get(a) = %d, Len = %d; want 2, 1", v, c.Len())
	}
}
