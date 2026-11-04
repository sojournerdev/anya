//go:build linux

package nvml

import (
	"context"
	"strings"
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/NVIDIA/go-nvml/pkg/nvml/mock/gpus"
	"github.com/NVIDIA/go-nvml/pkg/nvml/mock/server"
	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider"
	"github.com/sojournerdev/anya/internal/provider/contracttest"
)

func TestProviderSatisfiesContract(t *testing.T) {
	contracttest.Run(t, func(t *testing.T) provider.Provider {
		t.Helper()
		return newMockNVMLProvider(t)
	})
}

func TestProviderNormalizesMockNVMLMetrics(t *testing.T) {
	source := newMockNVMLProvider(t)

	observation, err := source.Observe(context.Background())
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if observation.Sequence != 1 {
		t.Fatalf("sequence = %d, want 1", observation.Sequence)
	}
	if observation.Identity.ID == "" || observation.Identity.Model == "" {
		t.Fatalf("identity = %#v, want populated identity", observation.Identity)
	}
	if observation.Performance.PowerWatts.Value != 50 {
		t.Fatalf("power = %v, want 50 watts", observation.Performance.PowerWatts.Value)
	}
	if observation.Memory.UsedBytes.Value != 40 {
		t.Fatalf("used memory = %d, want 40 bytes", observation.Memory.UsedBytes.Value)
	}
	if !observation.Capabilities.ECCSupported.Value {
		t.Fatal("ECCSupported = false, want true")
	}
}

func TestNewRejectsUnavailableGPUEnvironment(t *testing.T) {
	failureCases := []struct {
		name       string
		configure  func(*server.Server)
		wantPhrase string
	}{
		{
			name: "initialization failure",
			configure: func(api *server.Server) {
				api.InitFunc = func() nvml.Return { return nvml.ERROR_DRIVER_NOT_LOADED }
			},
			wantPhrase: "initialize NVML",
		},
		{
			name: "device count failure",
			configure: func(api *server.Server) {
				api.DeviceGetCountFunc = func() (int, nvml.Return) {
					return 0, nvml.ERROR_UNKNOWN
				}
			},
			wantPhrase: "get GPU count",
		},
		{
			name: "no GPU",
			configure: func(api *server.Server) {
				api.DeviceGetCountFunc = func() (int, nvml.Return) {
					return 0, nvml.SUCCESS
				}
			},
			wantPhrase: "no NVIDIA GPUs detected",
		},
		{
			name: "GPU selection failure",
			configure: func(api *server.Server) {
				api.DeviceGetHandleByIndexFunc = func(int) (nvml.Device, nvml.Return) {
					return nil, nvml.ERROR_NOT_FOUND
				}
			},
			wantPhrase: "select GPU 0",
		},
		{
			name: "UUID failure",
			configure: func(api *server.Server) {
				device := api.Devices[0].(*server.Device)
				device.GetUUIDFunc = func() (string, nvml.Return) {
					return "", nvml.ERROR_NOT_FOUND
				}
			},
			wantPhrase: "read GPU UUID",
		},
		{
			name: "model failure",
			configure: func(api *server.Server) {
				device := api.Devices[0].(*server.Device)
				device.GetNameFunc = func() (string, nvml.Return) {
					return "", nvml.ERROR_NOT_FOUND
				}
			},
			wantPhrase: "read GPU name",
		},
	}

	for _, test := range failureCases {
		t.Run(test.name, func(t *testing.T) {
			api, err := server.New(server.WithGPUs(gpus.A30_PCIE_24GB))
			if err != nil {
				t.Fatalf("create mock NVML API: %v", err)
			}
			test.configure(api)

			if _, err := newProvider(api); err == nil || !strings.Contains(err.Error(), test.wantPhrase) {
				t.Fatalf("newProvider() error = %v, want phrase %q", err, test.wantPhrase)
			}
		})
	}
}

func TestProviderPreservesObservationWhenMetricsFail(t *testing.T) {
	source := newMockNVMLProvider(t)
	device := source.device.(*server.Device)
	device.GetPowerUsageFunc = func() (uint32, nvml.Return) {
		return 0, nvml.ERROR_NOT_READY
	}
	device.GetTemperatureFunc = func(nvml.TemperatureSensors) (uint32, nvml.Return) {
		return 0, nvml.ERROR_NOT_SUPPORTED
	}

	observation, err := source.Observe(context.Background())
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if observation.Performance.PowerWatts.Status != gpu.FactFailed {
		t.Fatalf("power status = %q, want %q", observation.Performance.PowerWatts.Status, gpu.FactFailed)
	}
	if !observation.Performance.PowerWatts.Error.Retryable {
		t.Fatal("power failure is not marked retryable")
	}
	if observation.Thermal.TemperatureCelsius.Status != gpu.FactUnsupported {
		t.Fatalf("temperature status = %q, want %q", observation.Thermal.TemperatureCelsius.Status, gpu.FactUnsupported)
	}
}

func newMockNVMLProvider(t *testing.T) *Provider {
	t.Helper()

	api, err := server.New(server.WithGPUs(gpus.A30_PCIE_24GB))
	if err != nil {
		t.Fatalf("create mock NVML API: %v", err)
	}
	device := api.Devices[0].(*server.Device)
	device.SetMockFuncs()
	device.GetEccModeFunc = func() (nvml.EnableState, nvml.EnableState, nvml.Return) {
		return nvml.FEATURE_ENABLED, nvml.FEATURE_ENABLED, nvml.SUCCESS
	}
	device.GetMemoryInfoFunc = func() (nvml.Memory, nvml.Return) {
		return nvml.Memory{Total: 100, Used: 40, Free: 60}, nvml.SUCCESS
	}
	device.GetUtilizationRatesFunc = func() (nvml.Utilization, nvml.Return) {
		return nvml.Utilization{Gpu: 25}, nvml.SUCCESS
	}
	device.GetPowerUsageFunc = func() (uint32, nvml.Return) {
		return 50000, nvml.SUCCESS
	}
	device.GetClockInfoFunc = func(nvml.ClockType) (uint32, nvml.Return) {
		return 1000, nvml.SUCCESS
	}
	device.GetTemperatureFunc = func(nvml.TemperatureSensors) (uint32, nvml.Return) {
		return 50, nvml.SUCCESS
	}
	device.GetTemperatureThresholdFunc = func(nvml.TemperatureThresholds) (uint32, nvml.Return) {
		return 90, nvml.SUCCESS
	}
	device.GetCurrentClocksThrottleReasonsFunc = func() (uint64, nvml.Return) {
		return 0, nvml.SUCCESS
	}

	provider, err := newProvider(api)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	t.Cleanup(func() {
		if err := provider.Close(); err != nil {
			t.Errorf("close provider: %v", err)
		}
	})
	return provider
}
