package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/sojournerdev/anya/internal/agent"
	"github.com/sojournerdev/anya/internal/app"
	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider/fake"
)

func TestNewRejectsInvalidApplicationConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		nodeAgent *agent.Agent
		interval  time.Duration
		timeout   time.Duration
	}{
		{name: "nil agent", interval: time.Second, timeout: time.Second},
		{name: "non-positive interval", nodeAgent: newTestAgent(t), timeout: time.Second},
		{name: "timeout exceeds interval", nodeAgent: newTestAgent(t), interval: time.Second, timeout: 2 * time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := app.New(
				app.Dependencies{Agent: test.nodeAgent},
				app.Settings{RefreshInterval: test.interval, RefreshTimeout: test.timeout, ListenAddress: ":0"},
			); err == nil {
				t.Fatal("New() succeeded, want error")
			}
		})
	}
}

func TestApplicationRunRetriesAfterRefreshFailure(t *testing.T) {
	firstObservation := validObservation()
	secondObservation := validObservation()
	secondObservation.Sequence = 2
	source := fake.New(
		fake.Sample{Observation: firstObservation},
		fake.Sample{Observation: secondObservation},
	)
	nodeAgent, err := agent.New(source)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	application := newTestApplication(t, nodeAgent, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- application.Run(ctx)
	}()

	deadline := time.After(time.Second)
	for {
		snapshot, ok := nodeAgent.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour}).Snapshot()
		if ok && snapshot.Sequence == 2 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("application did not perform periodic refresh")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after cancellation")
	}
}

func TestApplicationRunRetriesAfterInitialRefreshFailure(t *testing.T) {
	source := fake.New(
		fake.Sample{Observation: gpu.Observation{CapturedAt: time.Now()}},
		fake.Sample{Observation: validObservation()},
	)
	nodeAgent, err := agent.New(source)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	application := newTestApplication(t, nodeAgent, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- application.Run(ctx)
	}()

	deadline := time.After(time.Second)
	for {
		if _, ok := nodeAgent.Evaluate(time.Now(), agent.Policy{MaxStateAge: time.Hour}).Snapshot(); ok {
			break
		}
		select {
		case <-deadline:
			t.Fatal("application did not recover after initial refresh failure")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after cancellation")
	}
}
func TestApplicationRunStopsBeforeInitialRefreshWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	application := newTestApplication(t, newTestAgent(t), time.Second)
	if err := application.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
}

func TestApplicationRunStopsWhileWaitingWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := newTestApplication(t, newTestAgent(t), time.Hour)

	result := make(chan error, 1)
	go func() {
		result <- application.Run(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after cancellation")
	}
}

func newTestApplication(t *testing.T, nodeAgent *agent.Agent, refreshInterval time.Duration) *app.Application {
	t.Helper()
	httpServer, err := app.NewServer(nodeAgent, time.Second)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	application, err := app.New(
		app.Dependencies{
			Agent:      nodeAgent,
			HTTPServer: httpServer,
		},
		app.Settings{
			RefreshInterval: refreshInterval,
			RefreshTimeout:  time.Millisecond,
			ListenAddress:   ":0",
		},
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return application
}
func newTestAgent(t *testing.T) *agent.Agent {
	t.Helper()
	nodeAgent, err := agent.New(fake.New(fake.Sample{
		Observation: validObservation(),
	}))
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	return nodeAgent
}

func validObservation() gpu.Observation {
	unknownUint := gpu.Fact[uint64]{Status: gpu.FactUnknown}
	unknownFloat := gpu.Fact[float64]{Status: gpu.FactUnknown}
	unknownBool := gpu.Fact[bool]{Status: gpu.FactUnknown}
	return gpu.Observation{
		Sequence:   1,
		CapturedAt: time.Unix(1, 0),
		Identity:   gpu.Identity{ID: "test-gpu", Model: "test-model"},
		Capabilities: gpu.PlatformCapabilities{
			MemoryBytes:  unknownUint,
			ECCSupported: unknownBool,
		},
		Performance: gpu.Performance{
			UtilizationPercent: unknownFloat,
			PowerWatts:         unknownFloat,
			GraphicsClockMHz:   unknownUint,
			MemoryClockMHz:     unknownUint,
		},
		Memory: gpu.Memory{
			TotalBytes: unknownUint,
			UsedBytes:  unknownUint,
			FreeBytes:  unknownUint,
		},
		Thermal: gpu.Thermal{
			TemperatureCelsius:      unknownFloat,
			TemperatureLimitCelsius: unknownFloat,
			Throttling:              unknownBool,
		},
	}
}
