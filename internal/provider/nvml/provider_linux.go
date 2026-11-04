//go:build linux

// Package nvml provides the production NVML observation adapter.
package nvml

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider"
)

// Provider observes one NVIDIA GPU through NVML.
type Provider struct {
	mutex    sync.Mutex
	api      nvml.Interface
	device   nvml.Device
	identity gpu.Identity
	sequence uint64
	closed   bool
}

var _ provider.Provider = (*Provider)(nil)

// New initializes NVML and selects the first available GPU.
func New() (provider.Provider, error) {
	return newProvider(nvml.New())
}

func newProvider(api nvml.Interface) (*Provider, error) {
	if api == nil {
		return nil, errors.New("nvml: API is required")
	}
	if returnCode := api.Init(); returnCode != nvml.SUCCESS {
		return nil, nvmlError("initialize NVML", returnCode)
	}

	gpuCount, returnCode := api.DeviceGetCount()
	if returnCode != nvml.SUCCESS {
		_ = api.Shutdown()
		return nil, nvmlError("get GPU count", returnCode)
	}
	if gpuCount == 0 {
		_ = api.Shutdown()
		return nil, errors.New("nvml: no NVIDIA GPUs detected; verify GPU assignment and NVIDIA runtime configuration")
	}

	device, returnCode := api.DeviceGetHandleByIndex(0)
	if returnCode != nvml.SUCCESS {
		_ = api.Shutdown()
		return nil, nvmlError("select GPU 0", returnCode)
	}

	gpuUUID, returnCode := device.GetUUID()
	if returnCode != nvml.SUCCESS {
		_ = api.Shutdown()
		return nil, nvmlError("read GPU UUID", returnCode)
	}
	gpuModel, returnCode := device.GetName()
	if returnCode != nvml.SUCCESS {
		_ = api.Shutdown()
		return nil, nvmlError("read GPU name", returnCode)
	}
	identity := gpu.Identity{ID: gpuUUID, Model: gpuModel}

	return &Provider{api: api, device: device, identity: identity}, nil
}

// Observe returns one normalized NVML sample.
func (p *Provider) Observe(ctx context.Context) (gpu.Observation, error) {
	if err := ctx.Err(); err != nil {
		return gpu.Observation{}, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err := ctx.Err(); err != nil {
		return gpu.Observation{}, err
	}
	if p.closed {
		return gpu.Observation{}, errors.New("nvml: provider is closed")
	}
	p.sequence++
	return p.collectObservation(p.sequence), nil
}

// Close releases the NVML library. It is idempotent.
func (p *Provider) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	if returnCode := p.api.Shutdown(); returnCode != nvml.SUCCESS && returnCode != nvml.ERROR_UNINITIALIZED {
		return nvmlError("shutdown NVML", returnCode)
	}
	return nil
}

func (p *Provider) collectObservation(sequence uint64) gpu.Observation {
	memoryInfo, memoryStatus := p.device.GetMemoryInfo()
	_, _, eccStatus := p.device.GetEccMode()

	utilization, utilizationStatus := p.device.GetUtilizationRates()
	powerMilliwatts, powerStatus := p.device.GetPowerUsage()
	graphicsClockMHz, graphicsClockStatus := p.device.GetClockInfo(nvml.CLOCK_GRAPHICS)
	memoryClockMHz, memoryClockStatus := p.device.GetClockInfo(nvml.CLOCK_MEM)

	temperatureCelsius, temperatureStatus := p.device.GetTemperature(nvml.TEMPERATURE_GPU)
	temperatureLimitCelsius, temperatureLimitStatus := p.device.GetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_GPU_MAX)
	throttleReasons, throttleStatus := p.device.GetCurrentClocksThrottleReasons()

	return gpu.Observation{
		Sequence:   sequence,
		CapturedAt: time.Now().UTC(),
		Identity:   p.identity,
		Capabilities: gpu.PlatformCapabilities{
			MemoryBytes:  uintFact(memoryInfo.Total, memoryStatus),
			ECCSupported: boolFact(true, eccStatus),
		},
		Performance: gpu.Performance{
			UtilizationPercent: floatFact(float64(utilization.Gpu), utilizationStatus),
			PowerWatts:         floatFact(float64(powerMilliwatts)/1000, powerStatus),
			GraphicsClockMHz:   uintFact(uint64(graphicsClockMHz), graphicsClockStatus),
			MemoryClockMHz:     uintFact(uint64(memoryClockMHz), memoryClockStatus),
		},
		Memory: gpu.Memory{
			TotalBytes: uintFact(memoryInfo.Total, memoryStatus),
			UsedBytes:  uintFact(memoryInfo.Used, memoryStatus),
			FreeBytes:  uintFact(memoryInfo.Free, memoryStatus),
		},
		Thermal: gpu.Thermal{
			TemperatureCelsius:      floatFact(float64(temperatureCelsius), temperatureStatus),
			TemperatureLimitCelsius: floatFact(float64(temperatureLimitCelsius), temperatureLimitStatus),
			Throttling:              boolFact(throttleReasons != 0, throttleStatus),
		},
	}
}

func nvmlError(operation string, returnCode nvml.Return) error {
	message := fmt.Sprintf("nvml: %s: %s", operation, nvml.ErrorString(returnCode))
	switch nvml.ErrorString(returnCode) {
	case "ERROR_LIBRARY_NOT_FOUND":
		return fmt.Errorf("%s; verify that the NVIDIA driver and container runtime expose NVML", message)
	default:
		return errors.New(message)
	}
}

func uintFact(value uint64, returnCode nvml.Return) gpu.Fact[uint64] {
	if returnCode == nvml.SUCCESS {
		return gpu.Fact[uint64]{Value: value, Status: gpu.FactAvailable}
	}
	return unavailableFact[uint64](returnCode)
}

func floatFact(value float64, returnCode nvml.Return) gpu.Fact[float64] {
	if returnCode == nvml.SUCCESS {
		return gpu.Fact[float64]{Value: value, Status: gpu.FactAvailable}
	}
	return unavailableFact[float64](returnCode)
}

func boolFact(value bool, returnCode nvml.Return) gpu.Fact[bool] {
	if returnCode == nvml.SUCCESS {
		return gpu.Fact[bool]{Value: value, Status: gpu.FactAvailable}
	}
	return unavailableFact[bool](returnCode)
}

func unavailableFact[T any](returnCode nvml.Return) gpu.Fact[T] {
	factStatus := gpu.FactFailed
	if returnCode == nvml.ERROR_NOT_SUPPORTED {
		factStatus = gpu.FactUnsupported
	}
	return gpu.Fact[T]{Status: factStatus, Error: gpu.ObservationError{
		Code:      fmt.Sprintf("nvml_%d", returnCode),
		Message:   nvml.ErrorString(returnCode),
		Retryable: returnCode == nvml.ERROR_NOT_READY || returnCode == nvml.ERROR_TIMEOUT,
	}}
}
