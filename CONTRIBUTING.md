# Contributing

## Prerequisites

Install or enable the following before contributing:

| Requirement | Used for |
| --- | --- |
| [Git](https://git-scm.com/) | Clone the repository and manage changes. |
| [VS Code](https://code.visualstudio.com/) with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) | Open the supported development environment. |
| [Docker Engine](https://docs.docker.com/engine/) with [Buildx](https://docs.docker.com/reference/cli/docker/buildx/) support | Build and smoke-test the container image. |
| [`make`](https://www.gnu.org/software/make/) | Run repository workflows. |

The Dev Container provides Go and repository tools. A physical NVIDIA GPU is not required for unit tests, replay development, or container smoke tests.

Kubernetes validation is optional for contributors. It requires [`kubectl`](https://kubernetes.io/docs/reference/kubectl/), access to an operator-selected cluster, and the credentials described in `docs/OPERATOR-GUIDE.md`.

## Development Environment

Use the Dev Container for Go development and the replay provider. The container does not manage Kubernetes clusters or host Docker resources.

## Development Workflow

1. Open the repository in its development container.
2. Run Anya directly with the replay provider.
3. Use the container workflow when validating the production image locally.
4. Make the smallest change that solves the problem.
5. Add or update behavior-focused tests when runtime behavior changes.
6. Run the complete quality gate before opening a change for review.
7. Update the relevant source-of-truth document when a user-visible contract changes.

Use these commands for the supported local workflow:

| Command | Purpose |
| --- | --- |
| `make dev-run` | Run Anya with the replay provider. |
| `make dev-container-smoke` | Build the image and smoke-test the replay container. |
| `make ci` | Run the complete quality gate. |

Kubernetes deployment validation uses the operator-owned or homelab cluster path documented in `docs/OPERATOR-GUIDE.md`.

## Quality Checks

Run `make ci` from the development container before opening a change for review. The quality gate checks formatting, module consistency, builds, static analysis, race-enabled tests, dependency vulnerabilities, and lint.

Use individual targets while iterating:

| Command | Purpose |
| --- | --- |
| `make test` | Run unit tests. |
| `make race` | Run tests with the race detector. |
| `make vet` | Run `go vet`. |
| `make fmt-check` | Check Go formatting. |
| `make mod-tidy-check` | Check module file consistency. |
| `make mod-verify` | Verify module checksums. |
| `make vuln` | Run dependency vulnerability checks. |
| `make lint` | Run the configured linters. |

## Commit Messages

Use the [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>[optional scope]: <description>
```

Use an imperative, concise description. Common types include:

| Type | Use |
| --- | --- |
| `feat` | Add user-visible functionality. |
| `fix` | Correct incorrect behavior. |
| `docs` | Change documentation only. |
| `test` | Add or update tests. |
| `refactor` | Change implementation without changing behavior. |
| `perf` | Improve performance. |
| `build` | Change build or packaging behavior. |
| `ci` | Change continuous integration workflows. |
| `chore` | Change maintenance tooling or repository metadata. |

Examples:

```text
feat(replay): add cycle mode
fix(api): reject stale observations
docs(operators): clarify image publishing
```

Mark breaking changes with `!` after the optional scope:

```text
feat(api)!: replace the state response
```

## Documentation Responsibilities

Update documentation when changing an API endpoint, configuration variable, deployment value, image contract, deployment workflow, failure mode, security setting, or operator procedure.

Use these sources of truth:

- `README.md`: project context and navigation.
- `docs/DEVELOPMENT.md`: local development workflow.
- `docs/API-REFERENCE.md`: HTTP contract for consumers.
- `docs/OPERATOR-GUIDE.md`: deployment and production operations.
- `CONTRIBUTING.md`: repository contribution workflow.

## Pull Request Expectations

A change should state its behavioral effect, affected interfaces, and verification performed. Keep unrelated cleanup out of the same change. Do not add telemetry, retries, compatibility aliases, or new abstractions without a concrete requirement.

Document only workflows that exist. Label future workflows as `Planned`.
