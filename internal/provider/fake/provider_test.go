package fake_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider"
	"github.com/sojournerdev/anya/internal/provider/contracttest"
	"github.com/sojournerdev/anya/internal/provider/fake"
)

var errInjectedObservationFailure = errors.New("injected provider failure")

func TestProviderSatisfiesContract(t *testing.T) {
	contracttest.Run(t, func(t *testing.T) provider.Provider {
		t.Helper()
		return fake.New(
			fake.Sample{Observation: observationWithSequence(1)},
			fake.Sample{Observation: observationWithSequence(2)},
		)
	})
}

func TestProviderReturnsConfiguredObservationsInOrder(t *testing.T) {
	firstObservation := observationWithSequence(1)
	secondObservation := observationWithSequence(2)
	source := fake.New(
		fake.Sample{Observation: firstObservation},
		fake.Sample{Observation: secondObservation},
	)

	for _, wantObservation := range []gpu.Observation{firstObservation, secondObservation} {
		gotObservation, err := source.Observe(context.Background())
		if err != nil {
			t.Fatalf("Observe() error = %v", err)
		}
		if !reflect.DeepEqual(gotObservation, wantObservation) {
			t.Fatalf("observation = %#v, want %#v", gotObservation, wantObservation)
		}
	}
}

func TestProviderPropagatesConfiguredError(t *testing.T) {
	source := fake.New(fake.Sample{Err: errInjectedObservationFailure})

	_, err := source.Observe(context.Background())
	if err == nil {
		t.Fatal("Observe() succeeded, want error")
	}
	if !errors.Is(err, errInjectedObservationFailure) {
		t.Fatalf("Observe() error = %v, want injected observation failure", err)
	}
}

func TestProviderRejectsCanceledContext(t *testing.T) {
	source := fake.New(fake.Sample{Observation: observationWithSequence(1)})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := source.Observe(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Observe() error = %v, want context.Canceled", err)
	}
}

func TestProviderRejectsObservationAfterClose(t *testing.T) {
	source := fake.New(fake.Sample{Observation: observationWithSequence(1)})
	if err := source.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := source.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if _, err := source.Observe(context.Background()); err == nil {
		t.Fatal("Observe() after Close() succeeded")
	}
}

func observationWithSequence(sequence uint64) gpu.Observation {
	return gpu.Observation{Sequence: sequence}
}
