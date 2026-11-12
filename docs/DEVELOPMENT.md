# Development Guide

## Purpose

This guide defines the supported local development path for Anya.

## Local Development Workflow

Follow the shortest validation path that matches the change:

1. **Implement the change.**
2. **Run focused tests.** Use `make test` for package behavior; add `make race` when changing concurrency-sensitive code.
3. **Run Anya locally.** Install the Go version specified in [`go.mod`](../go.mod). The `make dev-run` target selects the replay provider:

   ```bash
   make dev-run
   ```

   This requires no GPU, image, registry, or Kubernetes cluster.
4. **Validate the production image when container behavior is affected:**

   ```bash
   make dev-container-smoke
   ```

   This builds the image, runs it with `AGENT_PROVIDER=replay`, and checks `/readyz`, `/healthz`, and `/v1/state`. It uses a dynamically allocated host port.
5. **Inspect the replay container manually when needed:**

   ```bash
   make dev-container-run
   curl http://localhost:8080/v1/state
   ```

   Override the host port with `make dev-container-run DEV_PORT=9090`.
6. **Run the complete quality gate before submitting:**

   ```bash
   make ci
   ```

## Edge Deployment Workflow

The edge workflow is optional local validation. It publishes a versioned image tagged with the current commit to the configured OCI registry, deploys the Kustomize overlay to an operator-selected Kubernetes cluster, waits for rollout, and verifies the service:

```bash
make edge-up
```

Use this only when the change requires edge or Kubernetes validation. The development host must provide Docker, `kubectl`, access to the configured OCI registry, and a configured Kubernetes context and image repository in `.env`. The cluster must be able to pull the published image and satisfy the runtime requirements in [`docs/OPERATOR-GUIDE.md`](OPERATOR-GUIDE.md).

## Replay Provider

`make dev-run`, `make dev-container-run`, and `make dev-container-smoke` use `AGENT_PROVIDER=replay` through the `DEV_PROVIDER` Make variable. Override it only when validating another provider:

```bash
make dev-run DEV_PROVIDER=nvml
```

The replay provider reads the embedded `healthy` fixture and cycles through its observations. It assigns monotonically increasing sequence numbers as observations are returned, so repeated runs provide a deterministic provider path without requiring NVIDIA libraries, a physical GPU, or GPU hardware access.

Use replay for:

- fast local development with repeatable input;
- unit and integration tests that must run without GPU hardware;
- container smoke tests that validate startup and HTTP behavior;
- exercising health, readiness, and snapshot-serving behavior.

Replay does not validate NVML integration, NVIDIA drivers, container GPU access, or real device behavior. Use the edge workflow and the requirements in [`docs/OPERATOR-GUIDE.md`](OPERATOR-GUIDE.md) for those checks.

`AGENT_PROVIDER` defaults to `nvml`. Replay is a development and test provider; production deployments use the NVML provider.

## Container Smoke Test

The container smoke test checks that the production image:

- Starts without GPU access.
- Serves the readiness and health endpoints.
- Returns an available replay state.

The test removes the temporary container after success or failure.
