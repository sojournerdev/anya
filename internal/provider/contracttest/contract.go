// Package contracttest provides shared provider behavior tests.
package contracttest

import (
	"context"
	"errors"
	"github.com/sojournerdev/anya/internal/provider"
	"testing"
)

// Factory creates a provider for one isolated contract test.
type Factory func(t *testing.T) provider.Provider

// Run checks the behavior required by provider.Provider.
func Run(t *testing.T, newProvider Factory) {
	t.Helper()

	newSource := func(t *testing.T) provider.Provider {
		t.Helper()
		source := newProvider(t)
		t.Cleanup(func() {
			if err := source.Close(); err != nil {
				t.Errorf("cleanup Close(): %v", err)
			}
		})
		return source
	}

	t.Run("returns increasing observations", func(t *testing.T) {
		source := newSource(t)
		first, err := source.Observe(context.Background())
		if err != nil {
			t.Fatalf("first Observe() error: %v", err)
		}
		second, err := source.Observe(context.Background())
		if err != nil {
			t.Fatalf("second Observe() error: %v", err)
		}
		if second.Sequence <= first.Sequence {
			t.Fatalf("sequences = %d, %d; want increasing values", first.Sequence, second.Sequence)
		}
	})

	t.Run("honors canceled context", func(t *testing.T) {
		source := newSource(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := source.Observe(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Observe() error = %v, want context.Canceled", err)
		}
	})

	t.Run("rejects observation after close", func(t *testing.T) {
		source := newSource(t)
		if err := source.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
		if _, err := source.Observe(context.Background()); err == nil {
			t.Fatal("Observe() succeeded after Close()")
		}
	})

	t.Run("close is idempotent", func(t *testing.T) {
		source := newSource(t)
		if err := source.Close(); err != nil {
			t.Fatalf("first Close() error: %v", err)
		}
		if err := source.Close(); err != nil {
			t.Fatalf("second Close() error: %v", err)
		}
	})

}
