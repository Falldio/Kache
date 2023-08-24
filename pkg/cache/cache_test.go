package cache

import "testing"

func TestBaseCacheBytes(t *testing.T) {
	c := &baseCache{}
	if c.Bytes() != 0 {
		t.Fatalf("expect 0, got %d", c.Bytes())
	}
}
