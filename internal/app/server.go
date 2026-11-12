package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sojournerdev/anya/internal/agent"
	"github.com/sojournerdev/anya/internal/health"
)

// NewServer creates the read-only HTTP server for agent state and probes.
func NewServer(nodeAgent *agent.Agent, policy agent.Policy) (*http.Server, error) {
	if nodeAgent == nil {
		return nil, errors.New("app: agent is required")
	}
	if policy.MaxStateAge <= 0 {
		return nil, errors.New("app: maximum state age must be positive")
	}

	return &http.Server{
		Handler:           newHTTPRoutes(nodeAgent, policy),
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}, nil
}

type agentStateHandler struct {
	nodeAgent *agent.Agent
	policy    agent.Policy
}

func newHTTPRoutes(nodeAgent *agent.Agent, policy agent.Policy) http.Handler {
	handler := agentStateHandler{
		nodeAgent: nodeAgent,
		policy:    policy,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.serveHealth)
	mux.HandleFunc("GET /readyz", handler.serveReadiness)
	mux.HandleFunc("GET /v1/state", handler.serveState)
	return mux
}

func (h agentStateHandler) serveHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(writer, http.StatusOK, probeResponse{Status: healthStatusOK})
}

func (h agentStateHandler) serveReadiness(writer http.ResponseWriter, _ *http.Request) {
	evaluation := h.nodeAgent.Evaluate(time.Now(), h.policy)
	if !evaluation.IsAvailable() {
		writeJSONResponse(
			writer,
			http.StatusServiceUnavailable,
			probeResponse{
				Status: readinessStatusNotReady,
				Code:   readinessCodeObservationUnavailable,
			},
		)
		return
	}
	writeJSONResponse(writer, http.StatusOK, probeResponse{Status: readinessStatusReady})
}

func (h agentStateHandler) serveState(writer http.ResponseWriter, _ *http.Request) {
	evaluation := h.nodeAgent.Evaluate(time.Now(), h.policy)
	if !evaluation.IsAvailable() {
		writeJSONResponse(
			writer,
			http.StatusServiceUnavailable,
			errorResponse{Error: stateError(evaluation.Reason())},
		)
		return
	}
	writeJSONResponse(writer, http.StatusOK, newStateResponse(evaluation))
}

const (
	healthStatusOK = "ok"
)

const (
	readinessStatusReady                = "ready"
	readinessStatusNotReady             = "not_ready"
	readinessCodeObservationUnavailable = "gpu_observation_unavailable"
)

const (
	stateErrorCodeUnavailable    = "state_unavailable"
	stateErrorMessageUnavailable = "GPU state is temporarily unavailable; the agent is retrying automatically"
	stateErrorCodeStale          = "state_stale"
	stateErrorMessageStale       = "GPU state is older than the configured maximum age"
)

func stateError(reason agent.Reason) apiError {
	if reason == agent.ReasonStaleSnapshot {
		return apiError{
			Code:    stateErrorCodeStale,
			Message: stateErrorMessageStale,
		}
	}
	return apiError{
		Code:    stateErrorCodeUnavailable,
		Message: stateErrorMessageUnavailable,
	}
}

type probeResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newStateResponse(evaluation agent.Evaluation) stateResponse {
	response := stateResponse{
		Availability: availabilityText(evaluation.IsAvailable()),
		Health:       healthText(evaluation.Health().Status),
		PublishedAt:  evaluation.PublishedAt(),
	}

	snapshot, hasSnapshot := evaluation.Snapshot()
	if !hasSnapshot {
		return response
	}

	response.Snapshot = &snapshotResponse{
		Sequence:   snapshot.Sequence,
		CapturedAt: snapshot.CapturedAt,
		Identity: identityView{
			ID:    snapshot.Identity.ID,
			Model: snapshot.Identity.Model,
		},
	}
	return response
}

type stateResponse struct {
	Availability string            `json:"availability"`
	Health       string            `json:"health"`
	PublishedAt  time.Time         `json:"published_at"`
	Snapshot     *snapshotResponse `json:"snapshot,omitempty"`
}

type snapshotResponse struct {
	Sequence   uint64       `json:"sequence"`
	CapturedAt time.Time    `json:"captured_at"`
	Identity   identityView `json:"identity"`
}

type identityView struct {
	ID    string `json:"id"`
	Model string `json:"model"`
}

func availabilityText(available bool) string {
	if available {
		return "available"
	}
	return "unknown"
}

func healthText(value health.Status) string {
	switch value {
	case health.StatusHealthy:
		return "healthy"
	case health.StatusDegraded:
		return "degraded"
	case health.StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

func writeJSONResponse(writer http.ResponseWriter, statusCode int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(value)
}
