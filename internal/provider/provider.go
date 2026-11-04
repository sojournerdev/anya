// Package provider defines the observation source contract.
package provider

import (
	"context"

	"github.com/sojournerdev/anya/internal/gpu"
)

// Kind identifies an observation provider implementation.
type Kind string

const (
	// KindNVML selects the NVIDIA Management Library provider.
	KindNVML Kind = "nvml"
	// KindReplay selects the deterministic replay provider.
	KindReplay Kind = "replay"
)

// Provider supplies normalized observations without creating agent state.
type Provider interface {

	// Observe returns one normalized sample or an error.
	Observe(ctx context.Context) (gpu.Observation, error)

	// Close releases provider resources.
	Close() error
}
