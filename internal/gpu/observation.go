package gpu

import "time"

// FactStatus records whether a provider fact can be used.
//
// Only FactAvailable makes Value meaningful. Other statuses explain why the
// value cannot be used.
type FactStatus string

// Fact status values.
const (
	FactUnknown     FactStatus = "Unknown"
	FactAvailable   FactStatus = "Available"
	FactUnsupported FactStatus = "Unsupported"
	FactFailed      FactStatus = "Failed"
)

// ObservationError describes why a provider fact is unavailable.
type ObservationError struct {
	Code      string
	Message   string
	Retryable bool
}

// PlatformCapabilities describes device limits and optional feature support.
type PlatformCapabilities struct {
	MemoryBytes  Fact[uint64]
	ECCSupported Fact[bool]
}

// Identity identifies a GPU and describes its model.
type Identity struct {
	ID    string
	Model string
}

// Performance contains instantaneous device measurements.
type Performance struct {
	UtilizationPercent Fact[float64]
	PowerWatts         Fact[float64]
	GraphicsClockMHz   Fact[uint64]
	MemoryClockMHz     Fact[uint64]
}

// Memory contains device memory measurements.
type Memory struct {
	TotalBytes Fact[uint64]
	UsedBytes  Fact[uint64]
	FreeBytes  Fact[uint64]
}

// Thermal contains temperature and throttling measurements.
type Thermal struct {
	TemperatureCelsius      Fact[float64]
	TemperatureLimitCelsius Fact[float64]
	Throttling              Fact[bool]
}

// Fact stores a value, its availability, and any provider error.
type Fact[T any] struct {
	Value  T
	Status FactStatus
	Error  ObservationError
}

// Observation is a normalized provider sample. It may be incomplete or invalid
// until NewSnapshot accepts it.
type Observation struct {
	Sequence     uint64
	CapturedAt   time.Time
	Identity     Identity
	Capabilities PlatformCapabilities
	Performance  Performance
	Memory       Memory
	Thermal      Thermal
}
