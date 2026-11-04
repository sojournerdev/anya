package health_test

import (
	"testing"
	"time"

	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/health"
)

func TestAssessValidSnapshotAsHealthyAtBaseline(t *testing.T) {
	snapshot, err := gpu.NewSnapshot(newValidObservation())
	if err != nil {
		t.Fatalf("NewSnapshot() error = %v", err)
	}

	assessment := health.Assess(snapshot)
	if assessment.Status != health.StatusHealthy {
		t.Fatalf("health status = %d, want %d", assessment.Status, health.StatusHealthy)
	}
	if len(assessment.Reasons) != 0 {
		t.Fatalf("health reasons = %#v, want none", assessment.Reasons)
	}
}

func newValidObservation() gpu.Observation {
	return gpu.Observation{
		Sequence:   1,
		CapturedAt: time.Unix(100, 0).UTC(),
		Identity:   gpu.Identity{ID: "test-gpu", Model: "test-model"},
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
