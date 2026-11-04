# API Reference

> **Implemented contract:** This document defines the consumer-facing HTTP contract implemented by the current build.

## Overview

Anya exposes a read-only HTTP API for health probes and the latest validated GPU state. The API describes the current node condition; it does not provide historical data or workload scheduling.

The default listen address is `:8080`. Kubernetes configures the application through environment values.

## Health Endpoint

```http
GET /healthz
```

The endpoint returns HTTP `200` when the HTTP process is serving requests:

```json
{"status":"ok"}
```

A successful response does not prove that a valid GPU observation exists.

## Readiness Endpoint

```http
GET /readyz
```

The endpoint returns HTTP `200` when Anya has published a fresh, valid GPU snapshot:
 
```json
{"status":"ready"}
```

Before the first successful observation, or after the current snapshot exceeds `AGENT_MAX_STATE_AGE`, it returns HTTP `503`:

```json
{"status":"not_ready","code":"gpu_observation_unavailable"}
```

Consumers should use readiness when deciding whether the node can serve GPU-aware work.

## State Endpoint

```http
GET /v1/state
```

When a fresh, valid snapshot exists, the endpoint returns HTTP `200`:

```json
{
  "availability": "available",
  "health": "healthy",
  "published_at": "2026-01-01T12:00:00Z",
  "snapshot": {
    "sequence": 42,
    "captured_at": "2026-01-01T11:59:59Z",
    "identity": {
      "id": "GPU-...",
      "model": "..."
    }
  }
}
```

If no valid snapshot has been published, the endpoint returns HTTP `503`:

```json
{
  "error": {
    "code": "state_unavailable",
    "message": "GPU state is temporarily unavailable; the agent is retrying automatically"
  }
}
```

If a snapshot exists but exceeds `AGENT_MAX_STATE_AGE`, the endpoint returns HTTP `503`:

```json
{
  "error": {
    "code": "state_stale",
    "message": "GPU state is older than the configured maximum age"
  }
}
```

The message describes the unavailable state. Process restart behavior belongs to the [Operator Guide](OPERATOR-GUIDE.md).

## Response Fields

| Field | Meaning |
| --- | --- |
| `availability` | `available` in every successful state response; the endpoint returns `503` when no valid or fresh snapshot exists. |
| `health` | The assessment of the published snapshot: `healthy`, `degraded`, `unhealthy`, or `unknown`. |
| `published_at` | The time Anya published the current snapshot. |
| `snapshot.sequence` | The observation sequence number. |
| `snapshot.captured_at` | The capture time reported by the GPU observation. |
| `snapshot.identity.id` | The GPU identifier. |
| `snapshot.identity.model` | The GPU model name. |

Consumers must handle health values other than `healthy`.

## Freshness and Sequence Semantics

Anya publishes a snapshot only after reading and validating an observation. A failed observation does not replace the last published state. An observation whose sequence is not newer than the current sequence is rejected. The published timestamp controls the maximum state age.

## HTTP Behavior

All endpoints are read-only. Responses use JSON and include `Cache-Control: no-store`.

The no-store directive prevents clients, proxies, and other HTTP intermediaries from retaining a response and serving an older snapshot after the agent has refreshed its state. `AGENT_REFRESH_INTERVAL` controls observation frequency and `AGENT_MAX_STATE_AGE` controls the maximum age of a published state.


## Compatibility

The state endpoint is versioned under `/v1`. Consumers should tolerate additional response fields and must not assume that `health: healthy` is the only valid health value.

Changes to endpoint paths, status codes, required fields, or field meanings require an API contract review before release.
