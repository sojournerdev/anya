# Observability Notes

## Purpose

The current signals show whether Anya is running, whether it has fresh GPU state, and why a refresh or state transition failed.

## Current Signals

The current build provides:

- Kubernetes liveness through `/healthz`.
- Kubernetes readiness through `/readyz`.
- State publication time through `published_at`.
- State freshness controlled by `AGENT_MAX_STATE_AGE`.
- Refresh failure messages in process logs.
- Pod restart and rollout status through Kubernetes.

The application does not currently expose a Prometheus `/metrics` endpoint or OpenTelemetry exporters.

## Freshness and Readiness

The refresh interval and maximum state age are separate settings:

- `AGENT_REFRESH_INTERVAL` controls observation frequency.
- `AGENT_MAX_STATE_AGE` controls the maximum age of a published state.

A failed observation does not replace the last valid snapshot. Anya continues retrying. Once the published state exceeds the maximum age, readiness fails and `/v1/state` returns `state_stale` until a fresh observation is published.

## Logs

Refresh failures are logged without credentials or registry data. Future structured logging should use bounded fields such as operation, result, and error category. Raw provider errors must not become metric labels.

## Operator Monitoring

The repository does not ship dashboards, alert rules, or monitoring integrations. Operators may monitor the following conditions using their own Kubernetes and logging systems:

- Unready pods.
- State age exceeding the configured limit.
- Repeated refresh failures.
- Restart loops.
- Failed image pulls.
- Failed rollouts.
- Provider initialization failures.

Thresholds and objectives are operator-defined; no availability or alerting objective is currently established by the repository.

## Future Telemetry

Prometheus metrics and OpenTelemetry are not currently implemented. Any future implementation must define exporter configuration, collector destination, sampling, resource cost, signal ownership, failure behavior, and bounded labels before operators depend on it.
