package replay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sojournerdev/anya/internal/provider"
	"github.com/sojournerdev/anya/internal/provider/contracttest"
)

func TestProviderSatisfiesContract(t *testing.T) {
	contracttest.Run(t, func(t *testing.T) provider.Provider {
		fixture, err := openEmbeddedFixture("healthy")
		if err != nil {
			t.Fatal(err)
		}
		source, err := newProvider(fixture, ModeOnce)
		_ = fixture.Close()
		if err != nil {
			t.Fatal(err)
		}
		return source
	})
}

func TestProviderHoldsFinalEntry(t *testing.T) {
	source := newTestProvider(t, `{"version":1,"observation":{"Sequence":1}}
{"version":1,"observation":{"Sequence":2}}
`, ModeHold)
	for want := uint64(1); want <= 2; want++ {
		observation, err := source.Observe(context.Background())
		if err != nil || observation.Sequence != want {
			t.Fatalf("Observe() = (%d, %v), want sequence %d", observation.Sequence, err, want)
		}
	}
	observation, err := source.Observe(context.Background())
	if err != nil || observation.Sequence != 3 {
		t.Fatalf("held Observe() = (%d, %v), want sequence 3", observation.Sequence, err)
	}
}

func TestProviderCyclesWithMonotonicSequences(t *testing.T) {
	source := newTestProvider(t, `{"version":1,"observation":{"Sequence":1}}
{"version":1,"observation":{"Sequence":2}}
`, ModeCycle)
	for want := uint64(1); want <= 6; want++ {
		observation, err := source.Observe(context.Background())
		if err != nil || observation.Sequence != want {
			t.Fatalf("Observe() = (%d, %v), want sequence %d", observation.Sequence, err, want)
		}
	}
}

func TestProviderReturnsTypedScriptedError(t *testing.T) {
	source := newTestProvider(t, `{"version":1,"error":{"code":"provider_unavailable","message":"GPU unavailable","retryable":true}}
`, ModeOnce)
	_, err := source.Observe(context.Background())
	var replayErr *Error
	if !errors.As(err, &replayErr) || replayErr.Code != "provider_unavailable" || !replayErr.Retryable {
		t.Fatalf("Observe() error = %v, want typed retryable replay error", err)
	}
}

func TestProviderRejectsInvalidFixture(t *testing.T) {
	_, err := newProvider(strings.NewReader(`{"version":1,"observation":{},"error":{"code":"both","message":"both"}}
`), ModeOnce)
	if err == nil {
		t.Fatal("newProvider() succeeded with invalid fixture")
	}
}

func TestProviderRejectsUnsupportedFixtureVersion(t *testing.T) {
	_, err := newProvider(strings.NewReader(`{"version":2,"observation":{"Sequence":1}}
`), ModeOnce)
	if err == nil {
		t.Fatal("newProvider() succeeded with unsupported version")
	}
}

func TestProviderRejectsCanceledContext(t *testing.T) {
	source := newTestProvider(t, `{"version":1,"observation":{"Sequence":1}}
`, ModeOnce)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := source.Observe(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Observe() error = %v, want context.Canceled", err)
	}
}

func newTestProvider(t *testing.T, content string, mode Mode) *Provider {
	t.Helper()
	source, err := newProvider(strings.NewReader(content), mode)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	return source
}
