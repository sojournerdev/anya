# Observability Notes

## Purpose

The current signals show whether Anya is running, whether it has fresh GPU state, and why a refresh or state transition failed.

## Current Signals

The current build provides:

- Kubernetes liveness through `/healthz`.
- Kubernetes readiness through `/readyz`.
- State publication time through `published_at`.
- Refresh failure messages in process logs.
- Pod restart and rollout status through Kubernetes.

The application does not currently expose a Prometheus `/metrics` endpoint or OpenTelemetry exporters.

## Freshness and Readiness

`AGENT_REFRESH_INTERVAL` controls observation frequency; `AGENT_MAX_STATE_AGE` sets the freshness limit. For snapshot publication and freshness semantics, see the [API Semantics](api-semantics.md) guide.

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

