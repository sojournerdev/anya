package factory_test

import (
	"testing"

	"github.com/sojournerdev/anya/internal/provider/factory"
)

func TestNewCreatesEmbeddedReplayProvider(t *testing.T) {
	source, err := factory.New("replay")
	if err != nil {
		t.Fatalf("factory.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Errorf("source.Close() error = %v", err)
		}
	})

	observation, err := source.Observe(t.Context())
	if err != nil {
		t.Fatalf("Observe() error = %v", err)
	}
	if observation.Sequence != 1 {
		t.Fatalf("sequence = %d, want 1", observation.Sequence)
	}
}

func TestNewRejectsUnknownProvider(t *testing.T) {
	_, err := factory.New("unknown")
	if err == nil {
		t.Fatal("factory.New() succeeded with unknown provider")
	}
}
