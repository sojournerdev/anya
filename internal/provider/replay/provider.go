// Package replay provides deterministic observation playback for local runs.
package replay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider"
)

const (
	fixtureVersion       = 1
	initialFixtureBuffer = 64 * 1024
	maxFixtureLine       = 1 << 20
)

// Mode controls behavior after the fixture reaches its final entry.
type Mode string

const (
	// ModeOnce stops after the final fixture entry.
	ModeOnce Mode = "once"

	// ModeHold repeats the final fixture entry.
	ModeHold Mode = "hold"

	// ModeCycle starts again at the first fixture entry.
	ModeCycle Mode = "cycle"
)

var (
	// ErrClosed indicates that the replay provider has been closed.
	ErrClosed = errors.New("replay provider is closed")
	// ErrFixtureExhausted indicates that a once-mode fixture has no entries left.
	ErrFixtureExhausted = errors.New("replay fixture is exhausted")
)

// Error describes a scripted provider failure.
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// Provider replays validated or failed observation outcomes in a deterministic order.
type Provider struct {
	mu           sync.Mutex
	entries      []entry
	mode         Mode
	index        int
	nextSequence uint64
	closed       bool
}

type entry struct {
	Version     int              `json:"version"`
	Observation *gpu.Observation `json:"observation,omitempty"`
	Error       *Error           `json:"error,omitempty"`
}

var _ provider.Provider = (*Provider)(nil)

// New initializes the replay provider with the embedded healthy fixture.
func New() (provider.Provider, error) {
	fixture, err := openEmbeddedFixture("healthy")
	if err != nil {
		return nil, err
	}
	source, err := newProvider(fixture, ModeCycle)
	closeErr := fixture.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, fmt.Errorf("replay: close embedded fixture: %w", closeErr)
	}
	return source, nil
}

// Observe returns the next fixture outcome.
func (p *Provider) Observe(ctx context.Context) (gpu.Observation, error) {
	if err := ctx.Err(); err != nil {
		return gpu.Observation{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return gpu.Observation{}, err
	}
	if p.closed {
		return gpu.Observation{}, ErrClosed
	}

	currentEntry, err := p.currentEntry()
	if err != nil {
		return gpu.Observation{}, err
	}
	p.advanceCursor()
	if currentEntry.Error != nil {
		return gpu.Observation{}, currentEntry.Error
	}
	observation := *currentEntry.Observation
	observation.Sequence = p.nextSequence
	p.nextSequence++
	return observation, nil
}

// Close makes the provider unavailable. It is idempotent.
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

func newProvider(reader io.Reader, mode Mode) (*Provider, error) {
	if reader == nil {
		return nil, errors.New("replay: reader is required")
	}
	if err := validatePlaybackMode(mode); err != nil {
		return nil, err
	}
	entries, err := readEntries(reader)
	if err != nil {
		return nil, err
	}
	return &Provider{entries: entries, mode: mode, nextSequence: 1}, nil
}

func validatePlaybackMode(playbackMode Mode) error {
	switch playbackMode {
	case ModeOnce, ModeHold, ModeCycle:
		return nil
	default:
		return fmt.Errorf("replay: unsupported playback mode %q", playbackMode)
	}
}

func readEntries(reader io.Reader) ([]entry, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, initialFixtureBuffer), maxFixtureLine)

	var entries []entry
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var replayEntry entry
		if err := json.Unmarshal(line, &replayEntry); err != nil {
			return nil, fmt.Errorf("replay: decode fixture line %d: %w", lineNumber, err)
		}
		if replayEntry.Version != fixtureVersion {
			return nil, fmt.Errorf("replay: fixture line %d has unsupported version %d", lineNumber, replayEntry.Version)
		}
		if (replayEntry.Observation == nil) == (replayEntry.Error == nil) {
			return nil, fmt.Errorf("replay: line %d must contain exactly one observation or error", lineNumber)
		}
		if replayEntry.Error != nil && (strings.TrimSpace(replayEntry.Error.Code) == "" || strings.TrimSpace(replayEntry.Error.Message) == "") {
			return nil, fmt.Errorf("replay: line %d error requires code and message", lineNumber)
		}

		entries = append(entries, replayEntry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("replay: read fixture: %w", err)
	}
	if len(entries) == 0 {
		return nil, errors.New("replay: fixture contains no entries")
	}
	return entries, nil
}

func (p *Provider) currentEntry() (entry, error) {
	if p.index >= len(p.entries) {
		switch p.mode {
		case ModeOnce:
			return entry{}, fmt.Errorf(
				"replay: fixture exhausted after %d entries: %w",
				len(p.entries),
				ErrFixtureExhausted,
			)
		case ModeHold:
			p.index = len(p.entries) - 1
		case ModeCycle:
			p.index = 0
		}
	}
	return p.entries[p.index], nil
}

func (p *Provider) advanceCursor() {
	switch p.mode {
	case ModeOnce:
		p.index++
	case ModeHold:
		if p.index < len(p.entries)-1 {
			p.index++
		}
	case ModeCycle:
		p.index = (p.index + 1) % len(p.entries)
	}
}
