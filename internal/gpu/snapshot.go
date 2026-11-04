package gpu

import "time"

// Snapshot is an immutable set of parsed GPU resources.
type Snapshot struct {
	sequence     uint64
	capturedAt   time.Time
	identity     Identity
	capabilities PlatformCapabilities
	performance  Performance
	memory       Memory
	thermal      Thermal
}

// NewSnapshot parses an observation and reports every issue it finds.
func NewSnapshot(observation Observation) (Snapshot, error) {
	var issues issueCollector
	metadata := buildMetadata(&issues, observation)

	snapshot := Snapshot{
		sequence:     metadata.sequence,
		capturedAt:   metadata.capturedAt,
		identity:     buildIdentity(&issues, observation.Identity),
		capabilities: buildCapabilities(&issues, observation.Capabilities),
		performance:  buildPerformance(&issues, observation.Performance),
		memory:       buildMemory(&issues, observation.Memory),
		thermal:      buildThermal(&issues, observation.Thermal),
	}

	if issues.invalid() {
		return Snapshot{}, &ParseError{issues: issues.result()}
	}
	return snapshot, nil
}

// Sequence returns the provider sequence number.
func (s Snapshot) Sequence() uint64 { return s.sequence }

// CapturedAt returns the provider capture time.
func (s Snapshot) CapturedAt() time.Time { return s.capturedAt }

// Identity returns the parsed GPU identity.
func (s Snapshot) Identity() Identity { return s.identity }

// Capabilities returns the parsed platform capabilities.
func (s Snapshot) Capabilities() PlatformCapabilities { return s.capabilities }

// Performance returns the parsed performance measurements.
func (s Snapshot) Performance() Performance { return s.performance }

// Memory returns the parsed memory measurements.
func (s Snapshot) Memory() Memory { return s.memory }

// Thermal returns the parsed thermal measurements.
func (s Snapshot) Thermal() Thermal { return s.thermal }
