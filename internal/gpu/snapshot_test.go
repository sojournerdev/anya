package gpu_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sojournerdev/anya/internal/gpu"
)

func TestNewSnapshotAcceptsValidObservation(t *testing.T) {
	capturedAt := time.Unix(100, 0)
	snapshot, err := gpu.NewSnapshot(validObservation(capturedAt))
	if err != nil {
		t.Fatalf("NewSnapshot() error = %v", err)
	}
	if got := snapshot.Sequence(); got != 1 {
		t.Fatalf("snapshot sequence = %d, want 1", got)
	}
	if got := snapshot.CapturedAt(); !got.Equal(capturedAt) {
		t.Fatalf("snapshot captured time = %v, want %v", got, capturedAt)
	}
	if got := snapshot.Identity().ID; got != "gpu-1" {
		t.Fatalf("snapshot identity = %q, want gpu-1", got)
	}
}

func TestNewSnapshotPreservesThermalLimit(t *testing.T) {
	observation := validObservation(time.Unix(100, 0))
	observation.Thermal.TemperatureCelsius = gpu.Fact[float64]{Value: 61, Status: gpu.FactAvailable}
	observation.Thermal.TemperatureLimitCelsius = gpu.Fact[float64]{Value: 62, Status: gpu.FactAvailable}

	snapshot, err := gpu.NewSnapshot(observation)
	if err != nil {
		t.Fatalf("NewSnapshot() error = %v", err)
	}
	if got := snapshot.Thermal().TemperatureLimitCelsius.Value; got != 62 {
		t.Fatalf("temperature limit = %v, want 62", got)
	}
}

func TestNewSnapshotRejectsTemperatureAboveLimit(t *testing.T) {
	observation := validObservation(time.Unix(100, 0))
	observation.Thermal.TemperatureCelsius = gpu.Fact[float64]{Value: 63, Status: gpu.FactAvailable}
	observation.Thermal.TemperatureLimitCelsius = gpu.Fact[float64]{Value: 62, Status: gpu.FactAvailable}

	_, err := gpu.NewSnapshot(observation)
	var parseErr *gpu.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("NewSnapshot() error = %v, want *gpu.ParseError", err)
	}
	if got := len(parseErr.Issues()); got != 1 {
		t.Fatalf("NewSnapshot() returned %d issues, want 1", got)
	}
}

func TestNewSnapshotReportsInvalidObservationIssues(t *testing.T) {
	observation := validObservation(time.Unix(100, 0))
	observation.Sequence = 0
	observation.Identity.ID = ""
	observation.Performance.UtilizationPercent = gpu.Fact[float64]{
		Value:  101,
		Status: gpu.FactAvailable,
	}

	_, err := gpu.NewSnapshot(observation)
	var parseErr *gpu.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("NewSnapshot() error = %v, want *gpu.ParseError", err)
	}
	if got := len(parseErr.Issues()); got != 3 {
		t.Fatalf("NewSnapshot() returned %d issues, want 3", got)
	}
}

func TestNewSnapshotRejectsContradictoryMemoryAccounting(t *testing.T) {
	observation := validObservation(time.Unix(100, 0))
	observation.Memory.TotalBytes = gpu.Fact[uint64]{Value: 100, Status: gpu.FactAvailable}
	observation.Memory.UsedBytes = gpu.Fact[uint64]{Value: 60, Status: gpu.FactAvailable}
	observation.Memory.FreeBytes = gpu.Fact[uint64]{Value: 60, Status: gpu.FactAvailable}

	_, err := gpu.NewSnapshot(observation)
	var parseErr *gpu.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("NewSnapshot() error = %v, want *gpu.ParseError", err)
	}
	if got := len(parseErr.Issues()); got != 1 {
		t.Fatalf("NewSnapshot() returned %d issues, want 1", got)
	}
}

func validObservation(capturedAt time.Time) gpu.Observation {
	return gpu.Observation{
		Sequence:   1,
		CapturedAt: capturedAt,
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
