package aop

import (
	"context"
	"testing"
)

func TestChain_Execute(t *testing.T) {
	var called []string

	i1 := InterceptorFunc(func(inv *Invocation) error {
		called = append(called, "before")
		err := inv.Proceed()
		called = append(called, "after")
		return err
	})

	chain := NewChain(i1)

	target := func(inv *Invocation) error {
		called = append(called, "target")
		return nil
	}

	_, err := chain.Execute(context.Background(), "test", target)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	expected := []string{"before", "target", "after"}
	if len(called) != len(expected) {
		t.Fatalf("Expected %d calls, got %d", len(expected), len(called))
	}
}

func TestWrap(t *testing.T) {
	var called bool

	i := InterceptorFunc(func(inv *Invocation) error {
		called = true
		return inv.Proceed()
	})

	chain := NewChain(i)

	handler := func(ctx context.Context, req string) (string, error) {
		return "response", nil
	}

	wrapped := Wrap(chain, "test", handler)

	resp, err := wrapped(context.Background(), "request")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if resp != "response" {
		t.Errorf("Expected 'response', got '%s'", resp)
	}

	if !called {
		t.Error("Interceptor not called")
	}
}
