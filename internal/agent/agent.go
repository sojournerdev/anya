// Package agent validates provider observations and publishes the latest usable GPU state.
package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/health"
	"github.com/sojournerdev/anya/internal/provider"
)

var (
	// ErrClosed is returned when a refresh is attempted after shutdown begins.
	ErrClosed = errors.New("agent: service is closed")
	// ErrStaleObservation is returned when an observation cannot advance the published sequence.
	ErrStaleObservation = errors.New("agent: observation sequence is not newer")
)

// Policy defines the freshness rules applied by Evaluate.
type Policy struct {
	// MaxStateAge is the longest period a published snapshot may be considered available.
	MaxStateAge time.Duration
}

// Agent coordinates provider refreshes and publishes the latest validated snapshot.
type Agent struct {
	provider    provider.Provider
	operationMu sync.Mutex
	stateMu     sync.RWMutex
	state       state
	closed      bool
	closeErr    error
}

// state is the agent's published value, including the snapshot health assessment
// and the time the agent accepted the snapshot.
type state struct {
	snapshot    gpu.Snapshot
	hasSnapshot bool
	health      health.Result
	publishedAt time.Time
}

// Reason explains the evaluator's availability decision.
type Reason string

const (
	// ReasonFreshSnapshot means a validated snapshot is within the freshness policy.
	ReasonFreshSnapshot Reason = "fresh_snapshot"
	// ReasonNoSnapshot means no validated snapshot has been published.
	ReasonNoSnapshot Reason = "no_snapshot"
	// ReasonStaleSnapshot means the latest snapshot is older than the freshness policy.
	ReasonStaleSnapshot Reason = "stale_snapshot"
)

// SnapshotView is the consumer-facing view of a published snapshot.
type SnapshotView struct {
	Sequence   uint64
	CapturedAt time.Time
	Identity   IdentityView
}

// IdentityView contains the stable identity fields exposed by SnapshotView.
type IdentityView struct {
	ID    string
	Model string
}

// Evaluation is the decision produced by applying Policy to the latest published state.
type Evaluation struct {
	state  state
	reason Reason
}

// IsAvailable reports whether the evaluation contains a fresh validated snapshot.
func (e Evaluation) IsAvailable() bool {
	return e.reason == ReasonFreshSnapshot
}

// Health returns the health assessment included in the evaluation.
func (e Evaluation) Health() health.Result { return e.state.health }

// Reason returns why the evaluator reached its availability decision.
func (e Evaluation) Reason() Reason { return e.reason }

// Snapshot returns the latest validated snapshot view, when one exists.
func (e Evaluation) Snapshot() (SnapshotView, bool) {
	if !e.state.hasSnapshot {
		return SnapshotView{}, false
	}
	snapshot := e.state.snapshot
	identity := snapshot.Identity()
	return SnapshotView{
		Sequence:   snapshot.Sequence(),
		CapturedAt: snapshot.CapturedAt(),
		Identity: IdentityView{
			ID:    identity.ID,
			Model: identity.Model,
		},
	}, true
}

// PublishedAt returns when the agent accepted the latest snapshot for publication.
func (e Evaluation) PublishedAt() time.Time { return e.state.publishedAt }

// New creates an agent that obtains observations from an observation provider.
func New(observationProvider provider.Provider) (*Agent, error) {
	if observationProvider == nil {
		return nil, errors.New("agent: provider is required")
	}
	return &Agent{provider: observationProvider}, nil
}

// Refresh obtains one provider observation, validates it, and publishes it when newer.
// Provider, parsing, and ordering failures leave the previously published state unchanged.
func (a *Agent) Refresh(ctx context.Context) error {
	if ctx == nil {
		return errors.New("agent: context is required")
	}

	a.operationMu.Lock()
	defer a.operationMu.Unlock()

	if a.closed {
		return ErrClosed
	}

	observation, err := a.provider.Observe(ctx)
	if err != nil {
		return fmt.Errorf("agent: observe provider: %w", err)
	}

	snapshot, err := gpu.NewSnapshot(observation)
	if err != nil {
		return fmt.Errorf("agent: parse snapshot: %w", err)
	}

	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	if a.state.hasSnapshot && snapshot.Sequence() <= a.state.snapshot.Sequence() {
		return ErrStaleObservation
	}
	a.state = stateFromSnapshot(snapshot, time.Now())
	return nil
}

// Evaluate applies policy to a consistent copy of the latest published state.
// It does not call the provider or mutate agent state.
func (a *Agent) Evaluate(now time.Time, policy Policy) Evaluation {
	a.stateMu.RLock()
	publishedState := a.state
	a.stateMu.RUnlock()

	return evaluatePublishedState(publishedState, now, policy)
}

func evaluatePublishedState(publishedState state, now time.Time, policy Policy) Evaluation {
	if !publishedState.hasSnapshot {
		return Evaluation{
			state:  publishedState.withUnknownHealth(),
			reason: ReasonNoSnapshot,
		}
	}
	if !publishedState.isFresh(now, policy.MaxStateAge) {
		return Evaluation{
			state:  publishedState.withUnknownHealth(),
			reason: ReasonStaleSnapshot,
		}
	}
	return Evaluation{
		state:  publishedState,
		reason: ReasonFreshSnapshot,
	}
}

// Close prevents future refreshes and releases provider resources.
func (a *Agent) Close() error {
	a.operationMu.Lock()
	defer a.operationMu.Unlock()

	if a.closed {
		return a.closeErr
	}
	a.closed = true
	a.closeErr = a.provider.Close()
	return a.closeErr
}

// withUnknownHealth returns a copy whose health cannot be established from current evidence.
func (s state) withUnknownHealth() state {
	s.health = health.Result{Status: health.StatusUnknown}
	return s
}

// isFresh reports whether the state was published within the configured age limit.
func (s state) isFresh(now time.Time, maxAge time.Duration) bool {
	if !s.hasSnapshot || maxAge <= 0 || s.publishedAt.IsZero() {
		return false
	}
	age := now.Sub(s.publishedAt)
	return age >= 0 && age <= maxAge
}

// stateFromSnapshot builds the complete state published after snapshot validation.
func stateFromSnapshot(snapshot gpu.Snapshot, now time.Time) state {
	assessment := health.Assess(snapshot)
	return state{
		snapshot:    snapshot,
		hasSnapshot: true,
		health:      assessment,
		publishedAt: now,
	}
}
