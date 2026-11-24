# Build and Image Distribution

## Scope

This document is the source of truth for Anya image builds, GitHub Container Registry distribution, image tags, and release maturity.

The project is in early development. Published images are development artifacts, not stable releases. API behavior, configuration, image contents, and Kubernetes resources may change without compatibility guarantees.

## Development Images

The image repository is:

```text
ghcr.io/sojournerdev/anya
```

GitHub Actions publishes both tags after required CI validation succeeds on pushes to `main`:

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

The image and Kustomize resources form a compatibility pair. See the [Operator Guide](operator-guide.md) for deployment and verification procedures.

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
