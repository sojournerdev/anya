package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sojournerdev/anya/internal/agent"
	"github.com/sojournerdev/anya/internal/app"
	"github.com/sojournerdev/anya/internal/gpu"
	"github.com/sojournerdev/anya/internal/provider/fake"
)

func TestNewServerRejectsNilAgent(t *testing.T) {
	if _, err := app.NewServer(nil, time.Second); err == nil {
		t.Fatal("NewServer() accepted a nil agent")
	}
}

func TestServerHealthEndpointReturnsJSONContract(t *testing.T) {
	_, server := newServerWithFakeAgent(t)

	response := serveRequest(server, http.MethodGet, "/healthz")
	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", response.Code, http.StatusOK)
	}
	assertJSONHeaders(t, response)

	var payload struct {
		Status string `json:"status"`
	}
	decodeJSON(t, response, &payload)
	if payload.Status != "ok" {
		t.Fatalf("health status body = %q, want ok", payload.Status)
	}
}

func TestServerHealthEndpointRejectsUnsupportedMethod(t *testing.T) {
	_, server := newServerWithFakeAgent(t)

	response := serveRequest(server, http.MethodPost, "/healthz")
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("health method status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", response.Header().Get("Allow"))
	}
}

func TestServerReadinessEndpointReturnsUnavailableBeforeRefresh(t *testing.T) {
	_, server := newServerWithFakeAgent(t)

	response := serveRequest(server, http.MethodGet, "/readyz")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	assertJSONHeaders(t, response)

	var payload struct {
		Status string `json:"status"`
		Code   string `json:"code"`
	}
	decodeJSON(t, response, &payload)
	if payload.Status != "not_ready" {
		t.Fatalf("readiness status body = %q, want not_ready", payload.Status)
	}
	if payload.Code != "gpu_observation_unavailable" {
		t.Fatalf("readiness code = %q, want gpu_observation_unavailable", payload.Code)
	}
}

func TestServerStateEndpointReturnsUnavailableBeforeRefresh(t *testing.T) {
	_, server := newServerWithFakeAgent(t)

	response := serveRequest(server, http.MethodGet, "/v1/state")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("state status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	assertJSONHeaders(t, response)

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeJSON(t, response, &payload)
	if payload.Error.Code != "state_unavailable" {
		t.Fatalf("state error code = %q, want state_unavailable", payload.Error.Code)
	}
}

func TestServerReadinessEndpointReturnsOKAfterRefresh(t *testing.T) {
	nodeAgent, server := newServerWithFakeAgent(t)
	if err := nodeAgent.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	response := serveRequest(server, http.MethodGet, "/readyz")
	if response.Code != http.StatusOK {
		t.Fatalf("readiness status = %d, want %d", response.Code, http.StatusOK)
	}
}
func TestServerReadinessEndpointReturnsUnavailableForStaleState(t *testing.T) {
	nodeAgent, server := newServerWithMaxAge(t, time.Nanosecond)
	if err := nodeAgent.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	time.Sleep(time.Millisecond)

	response := serveRequest(server, http.MethodGet, "/readyz")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestServerStateEndpointReturnsPublishedState(t *testing.T) {
	nodeAgent, server := newServerWithFakeAgent(t)
	if err := nodeAgent.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	response := serveRequest(server, http.MethodGet, "/v1/state")
	if response.Code != http.StatusOK {
		t.Fatalf("state status = %d, want %d", response.Code, http.StatusOK)
	}
	assertJSONHeaders(t, response)

	var payload struct {
		Availability string `json:"availability"`
		Health       string `json:"health"`
		Snapshot     struct {
			Sequence uint64 `json:"sequence"`
		} `json:"snapshot"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode state response: %v", err)
	}
	if payload.Availability != "available" {
		t.Fatalf("availability = %q, want available", payload.Availability)
	}
	if payload.Health != "healthy" {
		t.Fatalf("health = %q, want healthy", payload.Health)
	}
	if payload.Snapshot.Sequence != 1 {
		t.Fatalf("snapshot sequence = %d, want 1", payload.Snapshot.Sequence)
	}
}

func newServerWithFakeAgent(t *testing.T) (*agent.Agent, *http.Server) {
	return newServerWithMaxAge(t, time.Second)
}

func newServerWithMaxAge(t *testing.T, maxAge time.Duration) (*agent.Agent, *http.Server) {
	t.Helper()
	nodeAgent, err := agent.New(fake.New(fake.Sample{Observation: validGPUObservation()}))
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}
	server, err := app.NewServer(nodeAgent, maxAge)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	return nodeAgent, server
}

func serveRequest(server *http.Server, method, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	return response
}

func assertJSONHeaders(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", response.Header().Get("Cache-Control"))
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", response.Header().Get("X-Content-Type-Options"))
	}
}

func decodeJSON(t *testing.T, response *httptest.ResponseRecorder, value any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(value); err != nil {
		t.Fatalf("decode JSON response: %v", err)
	}
}

func validGPUObservation() gpu.Observation {
	unknownUint := gpu.Fact[uint64]{Status: gpu.FactUnknown}
	unknownFloat := gpu.Fact[float64]{Status: gpu.FactUnknown}
	unknownBool := gpu.Fact[bool]{Status: gpu.FactUnknown}
	return gpu.Observation{
		Sequence:   1,
		CapturedAt: time.Unix(1, 0).UTC(),
		Identity:   gpu.Identity{ID: "GPU-test", Model: "test-model"},
		Capabilities: gpu.PlatformCapabilities{
			MemoryBytes:  unknownUint,
			ECCSupported: unknownBool,
		},
		Performance: gpu.Performance{
			UtilizationPercent: unknownFloat,
			PowerWatts:         unknownFloat,
			GraphicsClockMHz:   unknownUint,
			MemoryClockMHz:     unknownUint,
		},
		Memory: gpu.Memory{
			TotalBytes: unknownUint,
			UsedBytes:  unknownUint,
			FreeBytes:  unknownUint,
		},
		Thermal: gpu.Thermal{
			TemperatureCelsius:      unknownFloat,
			TemperatureLimitCelsius: unknownFloat,
			Throttling:              unknownBool,
		},
	}
}
