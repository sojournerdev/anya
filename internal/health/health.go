// Package health evaluates validated GPU snapshots.
package health

import "github.com/sojournerdev/anya/internal/gpu"

// Status describes the condition supported by the available evidence.
type Status uint8

const (
	// StatusUnknown means the available evidence cannot determine health.
	StatusUnknown Status = iota
	// StatusHealthy means the validated snapshot meets the baseline.
	StatusHealthy
	// StatusDegraded means the snapshot indicates a non-critical condition.
	StatusDegraded
	// StatusUnhealthy means the snapshot indicates a critical condition.
	StatusUnhealthy
)

// Result is the health assessment for one snapshot.
type Result struct {
	Status  Status
	Reasons []Reason
}

// Reason explains an assessment result.
type Reason struct {
	Code    string
	Message string
}

// Assess returns baseline health for a validated GPU snapshot.
//
// A validated snapshot establishes that the provider returned usable GPU data.
// Detailed hardware checks can later produce degraded or unhealthy results.
func Assess(_ gpu.Snapshot) Result {
	return Result{Status: StatusHealthy}
}
