# Security Notes

## Security Boundary

Anya exposes a read-only decision API for trusted cluster consumers. The application does not provide authentication or authorization.

The repository does not ship a `NetworkPolicy`. Operators must provide network isolation if the service requires restricted access. Do not expose the service publicly without an approved authentication and transport-security design.

TLS termination, if required, occurs at an approved cluster boundary. The application process does not terminate TLS.

## Container Security

The Kubernetes DaemonSet configures Anya with:

- No service-account token.
- No privilege escalation.
- All Linux capabilities dropped.
- A read-only root filesystem.
- A non-root user (`65532:65532`).
- The runtime-default seccomp profile.
- Bounded CPU and memory requests and limits.

## Image and Build Security

The default edge repository is `ghcr.io/sojournerdev/anya`. The local edge workflow publishes an image tagged `git-<commit>` and deploys that image through a rendered Kustomize overlay. The base container images are pinned by digest.

CI runs `govulncheck` against Go dependencies. The repository does not currently generate or verify image signatures, SBOMs, provenance attestations, or registry image-scan results.

Operators are responsible for registry access, image-pull permissions, and any admission policy that requires signed or digest-pinned images.

## Configuration and Secrets

The documented configuration values are non-secret. Registry credentials belong in the operator's Kubernetes secret or image-pull mechanism and must not be committed to Kustomize resources or source control.

## Security Changes

Changes to container privileges, network access, image provenance, authentication, authorization, or secret handling require a security review and an update to this document.
