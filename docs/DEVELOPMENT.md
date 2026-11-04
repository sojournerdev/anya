# Development Guide

## Purpose

This guide defines the supported local development path for Anya.

## Local Development Workflow

The Dev Container provides Go and repository development tools. It does not manage Kubernetes clusters, Docker sockets, or host infrastructure.

Run Anya directly with the replay provider:

```bash
make dev-run
```

This is the fastest development loop. It requires no GPU, image, registry, or Kubernetes cluster.

Validate the production image from the host checkout because this workflow uses Docker:

```bash
make dev-container-smoke
```

This builds the image, runs it with `AGENT_PROVIDER=replay`, and checks `/readyz`, `/healthz`, and `/v1/state`.

Run the replay container on host port `8080` when you want to inspect it manually:

```bash
make dev-container-run
curl http://localhost:8080/v1/state
```

Override the host port when needed:

```bash
make dev-container-run DEV_PORT=9090
```

The smoke-test target uses a dynamically allocated host port and should be used when a fixed port is not required.

The host-side edge workflow publishes an immutable image to the configured OCI registry and deploys the Kustomize overlay to an operator-selected Kubernetes cluster:

```bash
make edge-up
```

The host must provide Docker, `kubectl`, and access to the configured OCI registry and Kubernetes context. Configure the deployment context and image repository in `.env`.

## Replay Provider

The application selects the provider with one environment variable:

```bash
AGENT_PROVIDER=replay
```

The replay provider uses the embedded `healthy` fixture and cycles through its observations. It does not require NVIDIA libraries or a GPU.

`AGENT_PROVIDER` defaults to `nvml`. The replay provider is a development and test tool. Production deployments use the NVML provider.

## Development Environment

The development environment provides pinned Go tooling and the commands used by the repository workflow. It does not require a physical NVIDIA GPU for unit tests, static checks, or replay-container validation.

GPU-provider behavior is tested with fakes and replay fixtures. A real NVIDIA GPU is required for the edge-node integration path.

## Quality Gate

Run the complete quality gate before submitting a change:

```bash
make ci
```

The gate checks formatting, module consistency, builds, static analysis, race-enabled tests, dependency vulnerabilities, and lint.

## Container Smoke Test

The container smoke test checks that the production image:

- Starts without GPU access.
- Serves the readiness and health endpoints.
- Returns an available replay state.

The test removes the temporary container after success or failure.

The edge workflow verifies the Kubernetes deployment against the operator-selected cluster.
