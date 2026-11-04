//go:build !linux

// Package nvml provides the production NVML observation adapter.
package nvml

import (
	"errors"

	"github.com/sojournerdev/anya/internal/provider"
)

// Provider preserves the NVML provider contract on non-Linux platforms.
type Provider struct{}

// New reports that NVML is unavailable on non-Linux platforms.
func New() (provider.Provider, error) {
	return nil, errors.New("nvml: provider requires Linux; run the agent on a Linux host with NVIDIA drivers")
}
