// Package fake provides deterministic provider behavior for tests.
package fake

import (
	"context"
	"errors"
	"sync"

	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider"
)

// Sample defines one configured provider outcome.
type Sample struct {
	Observation gpu.Observation
	Err         error
}

// Provider is a deterministic provider for tests and local runs.
type Provider struct {
	mu      sync.Mutex
	samples []Sample
	next    int
	closed  bool
}

var _ provider.Provider = (*Provider)(nil)

// New returns a provider configured with samples.
func New(samples ...Sample) *Provider {
	copiedSamples := append([]Sample(nil), samples...)
	return &Provider{samples: copiedSamples}
}

// Observe returns the next configured outcome.
func (p *Provider) Observe(ctx context.Context) (gpu.Observation, error) {
	if err := ctx.Err(); err != nil {
		return gpu.Observation{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return gpu.Observation{}, errors.New("fake provider: closed")
	}
	if p.next >= len(p.samples) {
		return gpu.Observation{}, errors.New("fake provider: no samples remaining")
	}

	sample := p.samples[p.next]
	p.next++
	return sample.Observation, sample.Err
}

// Close makes the provider unavailable. It is idempotent.
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}
