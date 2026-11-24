<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/anya-logo-on-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="docs/assets/anya-logo-on-light.svg">
    <img src="docs/assets/anya-logo-on-light.svg" alt="Anya logo" width="206" height="72">
  </picture>
</p>

<p align="center">A GPU State API for Single-GPU Edge Nodes</p>

<p align="center">
  <a href="https://github.com/sojournerdev/anya/actions/workflows/ci.yml"><img src="https://github.com/sojournerdev/anya/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI status"></a>
  <a href="https://github.com/sojournerdev/anya/blob/main/LICENSE"><img src="https://img.shields.io/github/license/sojournerdev/anya" alt="MIT license"></a>
  <a href="https://www.openapis.org/"><img src="https://img.shields.io/badge/OpenAPI-3.2.1-6BA539?logo=openapiinitiative" alt="OpenAPI 3.2.1"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/sojournerdev/anya" alt="Go version"></a>
</p>

Anya is a node agent that exposes a read-only HTTP API for validated NVIDIA GPU state on single-GPU edge nodes. Applications use this state to make hardware-aware decisions; workload and placement policies remain with consumers.

> [!WARNING]
> **Early development:** Anya is not production-ready. APIs, configuration, and
> deployment behavior may change without notice. Use only for development and
> testing.

## Goals

- Provide a small, well-documented API for GPU-aware consumers.
- Publish only validated GPU state.
- Expose Kubernetes-compatible liveness and readiness probes.
- Use minimal resources on a single-GPU edge node.
- Provide a repeatable Kubernetes deployment path.

## Non-goals

- Scheduling workloads.
- Managing GPU power, clocks, or driver settings.
- Replacing Kubernetes health or resource management.
- Providing long-term telemetry.
- Replacing [NVIDIA DCGM](https://developer.nvidia.com/dcgm).
- Aggregating multiple GPUs.
- Requiring operators to manage individual container commands for deployment.

## Roadmap

- [x] Publish validated GPU state through a read-only API
- [x] Support deterministic replay and container-based development
- [x] Provide a Kubernetes deployment path for single-GPU edge nodes
- [x] Add GitHub Actions quality checks
- [x] Define and publish the OpenAPI contract
- [ ] Add unified observability with structured logging, Prometheus metrics, and OpenTelemetry
- [ ] Expand GPU health and hardware awareness
- [ ] Assess PCIe and CUDA capabilities
- [ ] Support workload-specific hardware suitability decisions
- [ ] Validate on a Kubernetes cluster with NVIDIA GPU nodes
- [ ] Establish release and promotion practices

## Documentation

| Document | Audience | Purpose |
| --- | --- | --- |
| [API Contract](api/openapi.yaml) | API consumers and tooling | Canonical OpenAPI 3.2.1 contract for the HTTP API. |
| [API Semantics](docs/api-semantics.md) | API consumers | Snapshot lifecycle, availability and health, compatibility, and security behavior. |
| [Operator Guide](docs/operator-guide.md) | Operators and platform engineers | Kustomize and OCI registry deployment, configuration, health, troubleshooting, upgrades, and security. |
| [Development Guide](docs/development.md) | Contributors | Local development and API documentation workflow. |
| [Build and Image Distribution](docs/release.md) | Contributors and operators | Image distribution, development tags, release maturity, and supply-chain limitations. |
| [Security Notes](docs/security.md) | Operators and security reviewers | Implemented Kubernetes hardening, application boundary, and security limitations. |
| [Observability Notes](docs/observability.md) | Operators and developers | Current health, readiness, freshness, logs, and telemetry limitations. |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT License](LICENSE).
