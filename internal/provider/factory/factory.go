// Package factory constructs the configured observation provider.
package factory

import (
	"fmt"

	"github.com/sojournerdev/anya/internal/provider"
	"github.com/sojournerdev/anya/internal/provider/nvml"
	"github.com/sojournerdev/anya/internal/provider/replay"
)

// New constructs the provider selected by kind.
func New(providerKind provider.Kind) (provider.Provider, error) {
	switch providerKind {
	case provider.KindNVML:
		return nvml.New()
	case provider.KindReplay:
		return replay.New()
	default:
		return nil, fmt.Errorf("provider: unsupported provider %q", providerKind)
	}
}
