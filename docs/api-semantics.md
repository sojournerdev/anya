# API Semantics

The [OpenAPI contract](../api/openapi.yaml) defines the HTTP interface. This guide explains cross-cutting behavior and compatibility expectations that are clearer in prose.

## Snapshot lifecycle

Anya publishes a snapshot only after validating a provider observation. A failed observation does not replace the last published snapshot. New observations must have a strictly greater sequence number; duplicate or older observations are rejected.

`captured_at` is the timestamp reported by the provider. `published_at` is when Anya accepted the validated snapshot. Freshness is measured from `published_at`, not `captured_at`, and is limited by `AGENT_MAX_STATE_AGE`.

## Availability and health

Availability indicates whether Anya serves a fresh snapshot; `health` is a separate evaluator result. The current evaluator only validates snapshots; see the [OpenAPI contract](../api/openapi.yaml) for the meaning of `healthy`.

## Compatibility

The state endpoint is versioned under `/v1`. Consumers must tolerate additional response fields and should handle health values they do not recognize. Changes to paths, status codes, required fields, or field meaning require API contract review. The OpenAPI `info.version` identifies the contract version, not the application build version.

## Security boundary

The API does not provide authentication or authorization. Restrict network access to trusted clients; see the [Security Notes](security.md) and [Operator Guide](operator-guide.md) for deployment implications.
