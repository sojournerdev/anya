# Build and Image Distribution

## Scope

This document is the source of truth for Anya image builds, GitHub Container Registry distribution, image tags, and release maturity.

The project is in early development. Published images are development artifacts, not stable releases. API behavior, configuration, image contents, and Kubernetes resources may change without compatibility guarantees.

## Validation

Run the repository quality gate with:

```bash
make ci
```

For container behavior, run:

```bash
make dev-container-smoke
```

The smoke test uses the replay provider and does not require a GPU or Kubernetes cluster.

GitHub Actions runs these validations for pull requests and pushes to `main`. A successful push to `main` may publish the development image.

## Development Images

The image repository is:

```text
ghcr.io/sojournerdev/anya
```

Successful `main` builds receive both tags:

```text
ghcr.io/sojournerdev/anya:main
ghcr.io/sojournerdev/anya:sha-<commit>
```

- `main` is a moving tag for the newest successful development build.
- `sha-<commit>` identifies the source commit used to build a specific image.

Use `main` for disposable development environments. Use `sha-<commit>` for repeatable testing, issue reports, and rollback. Neither tag represents a stable release.

The image is currently published for `linux/amd64`.

## Kubernetes Deployment

The supported Kubernetes interface is the Kustomize overlay under `deploy/k8s/`. No Helm chart or automated release manifest is shipped.

The image and Kustomize resources form a compatibility pair. Operators should record the image tag and digest used with each deployment and verify rollout, readiness, and state behavior after changes.

The host-side edge workflow remains available for local or operator-managed validation:

```bash
make edge-up
```

It builds a local image, publishes it to the configured registry, applies the rendered Kustomize overlay, waits for the DaemonSet rollout, and verifies the service. It is not a stable release or promotion workflow.

## Future Versioned Releases

Semantic version tags will be introduced when the API and deployment contract are mature enough to support compatibility guarantees. Until then, the project does not publish stable semantic releases or maintain a promotion history.

## Rollback

Rollback is an operator procedure. Restore a previously tested commit-specific image and the corresponding Kustomize/resource revision, then verify:

- DaemonSet rollout
- Pod readiness
- State freshness
- Service/API behavior

## Supply-Chain Limitations

The repository currently does not generate or verify:

- Image signatures
- SBOMs
- Build provenance attestations
- Registry image-scan reports
- Generated release notes
- Cross-version Kubernetes, NVIDIA driver, runtime, or device-plugin compatibility reports
