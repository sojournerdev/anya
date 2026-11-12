# Anya

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/anya-logo-on-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="assets/anya-logo-on-light.svg">
    <img src="assets/anya-logo-on-light.svg" alt="Anya" width="206" height="72">
  </picture>
</p>

## A Local Decision API for Single-GPU Edge Nodes

Anya is a node agent for edge systems with one NVIDIA GPU. It publishes validated GPU state over HTTP so local consumers can make GPU-aware decisions without integrating directly with NVIDIA driver APIs.

Anya is designed for constrained edge nodes. It is not a scheduler, a GPU configuration manager, a multi-GPU aggregator, a telemetry store, or a replacement for [NVIDIA DCGM](https://developer.nvidia.com/dcgm).

> [!WARNING]
> **Early development:** Anya is not production-ready. APIs, configuration, and
> deployment behavior may change without notice. Use only for development,
> testing, and homelab validation.

## Goals

- Provide a small, stable API for GPU-aware consumers.
- Publish only validated GPU state.
- Expose Kubernetes-compatible liveness and readiness probes.
- Use minimal resources on a single-GPU edge node.
- Provide a repeatable Kubernetes deployment path.

## Non-goals

- Scheduling workloads.
- Managing GPU power, clocks, or driver settings.
- Replacing Kubernetes health or resource management.
- Providing long-term telemetry.
- Aggregating multiple GPUs.
- Requiring operators to manage individual container commands for deployment.

## Roadmap

- [x] Establish a validated GPU observation and decision API
- [x] Support deterministic replay and container-based development
- [x] Provide a Kubernetes deployment path for single-GPU edge nodes
- [x] Add GitHub Actions quality checks
- [ ] Define and publish the OpenAPI contract
- [ ] Add unified observability with structured logging, Prometheus metrics, and OpenTelemetry
- [ ] Expand GPU health and hardware awareness
- [ ] Assess PCIe and CUDA capabilities
- [ ] Support workload-specific hardware suitability decisions
- [ ] Validate behavior on real homelab GPU nodes
- [ ] Establish release and promotion practices

## Documentation

| Document | Audience | Purpose |
| --- | --- | --- |
| [API Reference](docs/API-REFERENCE.md) | API consumers | Implemented HTTP endpoints, response fields, errors, freshness, and compatibility. |
| [Operator Guide](docs/OPERATOR-GUIDE.md) | Operators and platform engineers | Kustomize and OCI registry deployment, configuration, health, troubleshooting, upgrades, and security. |
| [Development Guide](docs/DEVELOPMENT.md) | Contributors | Process and container development workflow. |
| [Build and Deployment Workflow](docs/RELEASE.md) | Contributors and operators | Local quality checks, container validation, optional edge deployment, and rollback limitations. |
| [Security Notes](docs/SECURITY.md) | Operators and security reviewers | Implemented Kubernetes hardening, application boundary, and security limitations. |
| [Observability Notes](docs/OBSERVABILITY.md) | Operators and developers | Current health, readiness, freshness, logs, and telemetry limitations. |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for repository workflow, quality checks, and documentation responsibilities.

## License

Anya is licensed under the [MIT License](LICENSE).
