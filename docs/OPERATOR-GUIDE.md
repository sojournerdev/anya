# Operator Guide

## Scope

Anya is a node agent for a single NVIDIA GPU on an edge node. It publishes validated GPU state over HTTP so trusted consumers can make GPU-aware decisions.

This guide describes how operators select, deploy, and verify a development image. Image tags and release maturity are defined in the [Build and Image Distribution](RELEASE.md) document.

## Ownership Model

The Anya project owns the container image, Kustomize resources, defaults, security settings, and compatibility contract.

Operators own the deployment decision and execution. Operators select image versions, provide cluster-specific configuration, manage registry access, perform upgrades, verify rollouts, and execute rollbacks.

## Deployment Golden Path

1. Select a published `sha-<commit>` image for repeatable validation.
2. Confirm the cluster can pull the image from GHCR.
3. Configure the Kubernetes context and image reference.
4. Apply the Kustomize overlay.
5. Wait for the DaemonSet rollout.
6. Verify health, readiness, and current state.
7. Record the deployed image tag or digest and rollout result.

The repository currently provides the Kustomize resources under `deploy/k8s/`. Helm is not currently a supported or shipped deployment interface.

## Prerequisites

The target deployment requires:

- A Linux node with a compatible NVIDIA GPU and driver.
- NVIDIA Container Toolkit support in the container runtime.
- Kubernetes with NVIDIA runtime support.
- The NVIDIA device plugin installed.
- A node label matching the deployment policy.
- `kubectl` configured for the target cluster.
- Network access to the configured OCI registry.
- Image-pull credentials when the registry package is private.

The deployment uses `kubectl` and Kustomize. The Makefile publishes images with Docker and applies the rendered overlay with `kubectl`.

## Image Selection

Use `ghcr.io/sojournerdev/anya:sha-<commit>` for repeatable cluster validation. The `main` tag is a moving development image and should be reserved for disposable environments. Do not use either tag as a stable production release.

Review [Build and Image Distribution](RELEASE.md) for the image policy. The image and Kustomize resources form a compatibility pair.

## Deployment

The supported host-side workflow for building and validating a local edge image is:

```bash
make edge-up
```

To deploy a CI-published image, provide its commit-specific tag and run the deployment stages:

```bash
make edge-deploy \
  KUBE_IMAGE=ghcr.io/sojournerdev/anya:sha-<commit>
make edge-rollout
make edge-verify
```

Individual stages are available:

```bash
make edge-build
make edge-publish
make edge-deploy
make edge-rollout
make edge-verify
```

The image repository and Kubernetes context are configured through `.env`:

```dotenv
KUBE_CONTEXT=your-context
EDGE_IMAGE_REPOSITORY=registry.example/anya
```

## Configuration

The application configuration includes:

| Setting | Default | Purpose |
| --- | --- | --- |
| `AGENT_PROVIDER` | `nvml` | Selects the production NVML provider. |
| `AGENT_HTTP_PORT` | `8080` | TCP IPv4 port used by the HTTP server. The agent listens on all pod/container interfaces. |
| `AGENT_REFRESH_INTERVAL` | `5s` | Time between observations after the initial observation. |
| `AGENT_REFRESH_TIMEOUT` | `2s` | Maximum time allowed for one provider observation. Must not exceed the refresh interval. |
| `AGENT_MAX_STATE_AGE` | `30s` | Maximum age of a published snapshot before readiness fails. Must be at least the refresh interval. |

Invalid configuration prevents startup. The process exits before provider or HTTP initialization.

## Health and Readiness

Use `/healthz` for liveness and `/readyz` for readiness. Liveness confirms that the HTTP process is serving requests. Readiness confirms that Anya has a fresh, valid GPU snapshot.

The pod remains unready until the first successful observation. A snapshot that exceeds `AGENT_MAX_STATE_AGE` is no longer served as usable state.


## Verification

After applying the Kustomize overlay, verify the workload:

```bash
kubectl --namespace anya-system get daemonset anya
kubectl --namespace anya-system get pods -l app.kubernetes.io/name=anya
kubectl --namespace anya-system rollout status daemonset/anya
```

Verify the service contract from an approved network location:

```bash
kubectl --namespace anya-system port-forward service/anya-agent 8080:8080
curl --fail http://127.0.0.1:8080/healthz
curl --fail http://127.0.0.1:8080/readyz
curl --fail http://127.0.0.1:8080/v1/state
```

The API reference defines the expected responses. Operators record the image tag or digest, rollout result, and verification result.

## Troubleshooting

### Pod is not ready

1. Check the DaemonSet and pod status.
2. Check pod events and logs.
3. Confirm that the NVIDIA device plugin assigned a GPU.
4. Confirm that the NVIDIA driver and container toolkit work on the node.
5. Query `/readyz` and `/v1/state`.
6. Check pod restart counts.

### Pod cannot start

Check the image reference, registry pull credentials, Kustomize overlay, node selection, NVIDIA runtime, and device-plugin status.

### State is stale

Compare `published_at` with the current time. A growing age usually indicates failed observations or pod restarts. Check logs, readiness, and restart counts together.

### Health is degraded or unhealthy

Treat `availability` and `health` separately. An available snapshot is not necessarily healthy. Consumers define policy for `degraded` and `unhealthy` states.

## Upgrades and Rollbacks

Before an upgrade:

1. Review the [Build and Image Distribution](RELEASE.md).
2. Confirm the image is available in the registry.
3. Confirm that the cluster can pull the image.
4. Review the Kustomize overlay and environment values.
5. Apply the updated overlay.
6. Wait for the DaemonSet rollout.
7. Verify health, readiness, and state.

If the rollout fails, inspect DaemonSet status, pod events, and logs. Restore the last known-good image and overlay, then verify readiness and state.

## Security and Resource Expectations

The application API has no authentication or authorization layer. The repository does not ship a NetworkPolicy; operators must provide network isolation if required.

See the [Security Notes](SECURITY.md) for the Kubernetes hardening controls and security limitations.

## Observability

See the [Observability Notes](OBSERVABILITY.md) for current health, readiness, freshness, logging, and telemetry limitations.

## Related Guides

- [API Reference](API-REFERENCE.md): consumer-facing HTTP contract.
- [Build and Image Distribution](RELEASE.md): image validation, GHCR distribution, development tags, and release maturity.
