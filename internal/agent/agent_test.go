package agent_test

import (
	"context"
	"errors"
	"github.com/sojournerdev/anya/internal/agent"
	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/health"
	"github.com/sojournerdev/anya/internal/provider/fake"
	"testing"
	"time"
)

func TestRefreshPublishesNewerStateAndPreservesLastStateAfterFailure(t *testing.T) {
	source := fake.New(
		fake.Sample{Observation: validObservation(1)},
		fake.Sample{Observation: validObservation(2)},
	)

	service, err := agent.New(source)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}

	initial := service.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour})
	if _, ok := initial.Snapshot(); ok {
		t.Fatal("Evaluate().Snapshot() reported state before the first successful refresh")
	}
	if initial.IsAvailable() || initial.Reason() != agent.ReasonNoSnapshot {
		t.Fatalf("initial evaluation = (available=%t, reason=%q), want unavailable/no_snapshot", initial.IsAvailable(), initial.Reason())
	}
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("first Refresh() error = %v", err)
	}

	state := service.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour})
	snapshot, ok := state.Snapshot()
	if !ok || snapshot.Sequence != 1 {
		t.Fatalf("first Evaluate().Snapshot() = (%#v, %t)", snapshot, ok)
	}
	if state.PublishedAt().IsZero() {
		t.Fatal("first state has zero update time")
	}
	if !state.IsAvailable() || state.Reason() != agent.ReasonFreshSnapshot {
		t.Fatalf("fresh evaluation = (available=%t, reason=%q), want available/fresh_snapshot", state.IsAvailable(), state.Reason())
	}
	if state.Health().Status != health.StatusHealthy {
		t.Fatalf("health = %v, want %v", state.Health().Status, health.StatusHealthy)
	}

	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("second Refresh() error = %v", err)
	}
	state = service.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour})
	snapshot, ok = state.Snapshot()
	if !ok || snapshot.Sequence != 2 {
		t.Fatalf("second Evaluate().Snapshot() = (%#v, %t)", snapshot, ok)
	}

	if err := service.Refresh(context.Background()); err == nil {
		t.Fatal("third Refresh() succeeded, want sample exhaustion error")
	}
	state = service.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour})
	snapshot, ok = state.Snapshot()
	if !ok || snapshot.Sequence != 2 {
		t.Fatalf("Evaluate().Snapshot() changed after failed refresh = (%#v, %t)", snapshot, ok)
	}
}

func TestEvaluateMarksExpiredStateUnavailable(t *testing.T) {
	service, err := agent.New(fake.New(
		fake.Sample{Observation: validObservation(1)},
	))
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	evaluation := service.Evaluate(
		time.Now().Add(time.Hour),
		agent.Policy{MaxStateAge: time.Minute},
	)
	if evaluation.IsAvailable() {
		t.Fatal("expired evaluation is available")
	}
	if evaluation.Reason() != agent.ReasonStaleSnapshot {
		t.Fatalf("evaluation reason = %q, want %q", evaluation.Reason(), agent.ReasonStaleSnapshot)
	}
	if evaluation.Health().Status != health.StatusUnknown {
		t.Fatalf("expired evaluation health = %v, want %v", evaluation.Health().Status, health.StatusUnknown)
	}
}

func TestRefreshRejectsInvalidObservationWithoutPublishing(t *testing.T) {
	source := fake.New(fake.Sample{Observation: gpu.Observation{CapturedAt: time.Now()}})
	service, err := agent.New(source)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}

	err = service.Refresh(context.Background())
	if _, ok := errors.AsType[*gpu.ParseError](err); !ok {
		t.Fatalf("Refresh() error = %v, want *gpu.ParseError", err)
	}
	if _, ok := service.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour}).Snapshot(); ok {
		t.Fatal("Evaluate().Snapshot() reported state after rejected observation")
	}
}

func TestNewRejectsNilProvider(t *testing.T) {
	if _, err := agent.New(nil); err == nil {
		t.Fatal("agent.New() accepted a nil provider")
	}
}

func TestRefreshRejectsAfterAgentClose(t *testing.T) {
	source := fake.New(fake.Sample{Observation: validObservation(1)})
	agentInstance, err := agent.New(source)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	if err := agentInstance.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := agentInstance.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if err := agentInstance.Refresh(context.Background()); err == nil {
		t.Fatal("Refresh() after Close() succeeded")
	}
}

func TestRefreshRejectsStaleObservation(t *testing.T) {
	source := fake.New(
		fake.Sample{Observation: validObservation(2)},
		fake.Sample{Observation: validObservation(1)},
	)
	service, err := agent.New(source)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("first Refresh() error = %v", err)
	}
	if err := service.Refresh(context.Background()); err == nil {
		t.Fatal("second Refresh() accepted a stale observation")
	}
	if snapshot, ok := service.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour}).Snapshot(); !ok || snapshot.Sequence != 2 {
		t.Fatalf("state changed after stale observation = (%#v, %t)", snapshot, ok)
	}
}

func validObservation(sequence uint64) gpu.Observation {
	return gpu.Observation{
		Sequence:   sequence,
		CapturedAt: time.Unix(1, 0),
		Identity:   gpu.Identity{ID: "gpu-1", Model: "test"},
		Capabilities: gpu.PlatformCapabilities{
			MemoryBytes:  gpu.Fact[uint64]{Status: gpu.FactUnknown},
			ECCSupported: gpu.Fact[bool]{Status: gpu.FactUnknown},
		},
		Performance: gpu.Performance{
			UtilizationPercent: gpu.Fact[float64]{Status: gpu.FactUnknown},
			PowerWatts:         gpu.Fact[float64]{Status: gpu.FactUnknown},
			GraphicsClockMHz:   gpu.Fact[uint64]{Status: gpu.FactUnknown},
			MemoryClockMHz:     gpu.Fact[uint64]{Status: gpu.FactUnknown},
		},
		Memory: gpu.Memory{
			TotalBytes: gpu.Fact[uint64]{Status: gpu.FactUnknown},
			UsedBytes:  gpu.Fact[uint64]{Status: gpu.FactUnknown},
			FreeBytes:  gpu.Fact[uint64]{Status: gpu.FactUnknown},
		},
		Thermal: gpu.Thermal{
			TemperatureCelsius:      gpu.Fact[float64]{Status: gpu.FactUnknown},
			TemperatureLimitCelsius: gpu.Fact[float64]{Status: gpu.FactUnknown},
			Throttling:              gpu.Fact[bool]{Status: gpu.FactUnknown},
		},
	}
}
