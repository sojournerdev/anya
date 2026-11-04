package gpu

import (
	"math"
	"strings"
	"time"
)

// IssueCode identifies why snapshot construction rejected an observation.
type IssueCode string

// Issue codes returned by snapshot construction.
const (
	IssueRequired      IssueCode = "required"
	IssueInvalidStatus IssueCode = "invalid_status"
	IssueMissingError  IssueCode = "missing_error"
	IssueOutOfRange    IssueCode = "out_of_range"
	IssueInconsistent  IssueCode = "inconsistent"
)

// ParseIssue describes one issue found while building a snapshot.
type ParseIssue struct {
	Path    string
	Code    IssueCode
	Message string
}

// String returns the issue path and human-readable message.
func (i ParseIssue) String() string {
	if i.Path == "" {
		return i.Message
	}
	return i.Path + ": " + i.Message
}

// ParseError reports issues found while building a snapshot.
type ParseError struct {
	issues []ParseIssue
}

// Error returns all parse issues in one message.
func (e *ParseError) Error() string {
	messages := make([]string, len(e.issues))
	for index, issue := range e.issues {
		messages[index] = issue.String()
	}
	return strings.Join(messages, "; ")
}

// Issues returns a copy of the issues that caused parsing to fail.
func (e *ParseError) Issues() []ParseIssue {
	return append([]ParseIssue(nil), e.issues...)
}

type issueCollector struct {
	issues []ParseIssue
}

func (c *issueCollector) add(path string, code IssueCode, message string) {
	c.issues = append(c.issues, ParseIssue{
		Path:    path,
		Code:    code,
		Message: message,
	})
}

func (c *issueCollector) invalid() bool {
	return len(c.issues) != 0
}

func (c *issueCollector) result() []ParseIssue {
	if len(c.issues) == 0 {
		return nil
	}
	return append([]ParseIssue(nil), c.issues...)
}

type snapshotMetadata struct {
	sequence   uint64
	capturedAt time.Time
}

func buildMetadata(c *issueCollector, observation Observation) snapshotMetadata {
	if observation.CapturedAt.IsZero() {
		c.add("observation.captured_at", IssueRequired, "capture time is required")
	}
	if observation.Sequence == 0 {
		c.add("observation.sequence", IssueRequired, "sequence must be greater than zero")
	}
	return snapshotMetadata{
		sequence:   observation.Sequence,
		capturedAt: observation.CapturedAt,
	}
}

func buildIdentity(c *issueCollector, input Identity) Identity {
	identity := Identity{
		ID:    strings.TrimSpace(input.ID),
		Model: strings.TrimSpace(input.Model),
	}
	if identity.ID == "" {
		c.add("identity.id", IssueRequired, "device ID is required")
	}
	return identity
}

func buildCapabilities(c *issueCollector, capabilities PlatformCapabilities) PlatformCapabilities {
	checkFactStatus(c, "capabilities.memory_bytes", capabilities.MemoryBytes)
	checkFactStatus(c, "capabilities.ecc_supported", capabilities.ECCSupported)
	return capabilities
}

func buildPerformance(c *issueCollector, performance Performance) Performance {
	checkFloatFact(c, "performance.utilization_percent", performance.UtilizationPercent, 0, 100)
	checkFloatFact(c, "performance.power_watts", performance.PowerWatts, 0, math.MaxFloat64)
	checkFactStatus(c, "performance.graphics_clock_mhz", performance.GraphicsClockMHz)
	checkFactStatus(c, "performance.memory_clock_mhz", performance.MemoryClockMHz)
	return performance
}

func buildMemory(c *issueCollector, memory Memory) Memory {
	checkFactStatus(c, "memory.total_bytes", memory.TotalBytes)
	checkFactStatus(c, "memory.used_bytes", memory.UsedBytes)
	checkFactStatus(c, "memory.free_bytes", memory.FreeBytes)
	if memory.TotalBytes.Status == FactAvailable && memory.UsedBytes.Status == FactAvailable && memory.UsedBytes.Value > memory.TotalBytes.Value {
		c.add("memory.used_bytes", IssueInconsistent, "used bytes exceed total bytes")
	}
	if memory.TotalBytes.Status == FactAvailable && memory.FreeBytes.Status == FactAvailable && memory.FreeBytes.Value > memory.TotalBytes.Value {
		c.add("memory.free_bytes", IssueInconsistent, "free bytes exceed total bytes")
	}
	if memory.TotalBytes.Status == FactAvailable &&
		memory.UsedBytes.Status == FactAvailable &&
		memory.FreeBytes.Status == FactAvailable &&
		memory.FreeBytes.Value <= memory.TotalBytes.Value &&
		memory.UsedBytes.Value > memory.TotalBytes.Value-memory.FreeBytes.Value {
		c.add("memory", IssueInconsistent, "used and free bytes exceed total bytes together")
	}
	return memory
}

func buildThermal(c *issueCollector, thermal Thermal) Thermal {
	checkFloatFact(c, "thermal.temperature_celsius", thermal.TemperatureCelsius, 0, math.MaxFloat64)
	checkFloatFact(c, "thermal.temperature_limit_celsius", thermal.TemperatureLimitCelsius, 0, math.MaxFloat64)
	if thermal.TemperatureCelsius.Status == FactAvailable &&
		thermal.TemperatureLimitCelsius.Status == FactAvailable &&
		thermal.TemperatureCelsius.Value > thermal.TemperatureLimitCelsius.Value {
		c.add("thermal.temperature_celsius", IssueInconsistent, "temperature exceeds reported limit")
	}
	checkFactStatus(c, "thermal.throttling", thermal.Throttling)
	return thermal
}

func checkFactStatus[T any](c *issueCollector, name string, fact Fact[T]) {
	switch fact.Status {
	case FactUnknown, FactAvailable, FactUnsupported:
	case FactFailed:
		if strings.TrimSpace(fact.Error.Code) == "" {
			c.add(name, IssueMissingError, "failed fact requires an error code")
		}
	default:
		c.add(name, IssueInvalidStatus, "invalid fact status")
	}
}

func checkFloatFact(c *issueCollector, name string, fact Fact[float64], minimum, maximum float64) {
	checkFactStatus(c, name, fact)
	if fact.Status != FactAvailable {
		return
	}
	if math.IsNaN(fact.Value) || math.IsInf(fact.Value, 0) {
		c.add(name, IssueOutOfRange, "value must be finite")
		return
	}
	if fact.Value < minimum || fact.Value > maximum {
		c.add(name, IssueOutOfRange, "value is outside the valid range")
	}
}
