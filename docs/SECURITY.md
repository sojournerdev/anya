# Security Guide

## Security Boundary

Anya exposes a read-only decision API for trusted cluster consumers. The API does not provide authentication or authorization in the application process.

The deployment restricts network access to required consumers and health probes. A NetworkPolicy or equivalent cluster control defines that boundary.

## Container Security

The deployment runs Anya with:

- No service-account token unless a documented feature requires one.
- No privilege escalation.
- All unnecessary Linux capabilities dropped.
- A read-only root filesystem.
- A non-root user.
- The runtime default seccomp profile.
- Bounded CPU and memory resources.

## Image Security

Production images come from GHCR and are referenced by immutable tag or digest. The release process produces an SBOM, vulnerability results, build provenance, and a signature.

The cluster verifies image-pull permissions. Signature verification is part of the deployment trust model when the cluster policy supports it.

## Network Exposure

The HTTP service is reachable only by the consumers and probes that require it. Operators do not expose the service publicly without an approved authentication and transport-security design.

TLS termination, if required, occurs at an approved cluster boundary. The API contract does not assume that the application process terminates TLS.

## Configuration and Secrets

Configuration values are non-secret unless explicitly documented otherwise. Registry credentials remain in Kubernetes secrets or the configured image-pull mechanism and do not appear in Kustomize resources committed to source control.

## Security Changes

Changes to container privileges, network access, image provenance, authentication, authorization, or secret handling require a security review and an update to this guide.
