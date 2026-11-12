# Build and Deployment Workflow

## Scope

The repository provides a local quality gate, container build, optional container smoke test, and optional Kubernetes edge deployment. It does not currently provide an automated release, promotion, signing, SBOM, or rollback pipeline.

## Quality Validation

Run the repository quality gate with:

```bash
make ci
```

This runs formatting checks, module checks, build, vet, race tests, `govulncheck`, and lint.

For container behavior, run:

```bash
make dev-container-smoke
```

The container smoke test uses the replay provider and does not require a GPU or Kubernetes cluster.

## Edge Workflow

The optional edge workflow is:

```bash
make edge-up
```

It builds an image, pushes it to the configured OCI registry, renders the Kustomize overlay with the image reference, applies it to the selected cluster, waits for the DaemonSet rollout, and verifies the service API through port-forwarding.

The default image naming convention is:

```text
ghcr.io/sojournerdev/anya:git-<commit>
```

The workflow requires Docker, `kubectl`, registry access, a configured Kubernetes context, and a cluster with the NVIDIA runtime and device plugin required by the DaemonSet.

## Deployment Interface

The supported Kubernetes interface is the Kustomize overlay under `deploy/k8s/`. No Helm chart or automated release manifest is shipped.

The deployment resources and image are treated as a compatibility pair. Changes to endpoint behavior, configuration, image behavior, or Kubernetes resources require the corresponding validation and documentation updates.

## Rollback

Rollback is an operator procedure. Restore a previously known image tag and Kustomize/resource revision, then verify:

- DaemonSet rollout
- Pod readiness
- State freshness
- Service/API behavior

The repository does not record approved releases, known-good digests, promotion history, or automated rollback metadata.

## Supply-Chain Limitations

The repository currently does not generate or verify:

- Image signatures
- SBOMs
- Build provenance attestations
- Registry image-scan reports
- Release notes as a generated artifact
- Cross-version Kubernetes, NVIDIA driver, runtime, or device-plugin compatibility reports
