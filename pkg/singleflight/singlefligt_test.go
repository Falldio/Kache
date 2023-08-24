package singleflight

import (
	"fmt"
	"testing"
	"time"
)

func TestGroup_Do(t *testing.T) {
	g := &Group{}

	// Test executing a function that returns a value
	v, err := g.Do("key", func() (interface{}, error) {
		return "value", nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if v != "value" {
		t.Errorf("expected value %q, got %q", "value", v)
	}

	// Test executing a function that returns an error
	v, err = g.Do("key", func() (interface{}, error) {
		return nil, fmt.Errorf("error")
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if v != nil {
		t.Errorf("expected nil value, got %v", v)
	}

	// Test executing a function concurrently
	go g.Do("key", func() (any, error) {
		time.Sleep(100 * time.Millisecond)
		return "value", nil
	})
	time.Sleep(50 * time.Millisecond)
	flag := false
	g.Do("key", func() (any, error) {
		flag = true
		return "value", nil
	})

	if flag {
		t.Error("expected function not to be called")
	}
}
