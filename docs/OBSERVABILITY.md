# Observability Guide

## Purpose

Observability shows whether Anya is running, whether it has fresh GPU state, and why a refresh or state transition failed.

## Current Signals

The current build provides:

- Kubernetes liveness through `/healthz`.
- Kubernetes readiness through `/readyz`.
- State publication time through `published_at` and state freshness through `AGENT_MAX_STATE_AGE`.
- Refresh failure messages in process logs.
- Pod restart and rollout signals from Kubernetes.

The application does not currently expose a Prometheus `/metrics` endpoint or OpenTelemetry exporters.

## Freshness and Readiness

The refresh interval and maximum state age are separate settings:

- `AGENT_REFRESH_INTERVAL` controls observation frequency.
- `AGENT_MAX_STATE_AGE` controls the maximum age of a published state.

A failed observation does not replace the last valid snapshot. Anya continues retrying. Once the published state exceeds the maximum age, readiness fails and `/v1/state` returns `state_stale` until a fresh observation is published.

## Logs

Current refresh failures are logged without credentials or registry data. Future structured logging should use bounded fields such as operation, result, and error category. Raw provider errors must not become metric labels.

## Planned Metrics

When metrics are added, the initial bounded signal set should cover:

- Refresh attempts and failures.
- Refresh duration.
- Current state age.
- Readiness state.
- HTTP request count, duration, and status class.

Metric labels must remain bounded. Do not use GPU identifiers, sequence numbers, raw error messages, timestamps, or arbitrary user-provided values as labels.

## OpenTelemetry

OpenTelemetry is not currently implemented. If added, the implementation must define exporter configuration, collector destination, sampling, resource cost, signal ownership, and failure behavior before operators depend on it.

## Alerts

The current operational alert targets are:

- Unavailable or unready pods.
- State age above the allowed threshold.
- Repeated refresh failures visible in logs.
- Restart loops.
- Failed image pulls.
- Failed rollouts.
- Provider initialization failures.

Alert thresholds must match the documented freshness and availability objectives.
