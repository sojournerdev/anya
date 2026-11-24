# Contributing

## Prerequisites

Install or enable the following before contributing:

| Requirement | Used for |
| --- | --- |
| [Git](https://git-scm.com/) | Clone the repository and manage changes. |
| [Go](https://go.dev/dl/) | Build, test, and run Anya locally. |
| [Docker Engine](https://docs.docker.com/engine/) with [Buildx](https://docs.docker.com/reference/cli/docker/buildx/) support | Build and smoke-test the container image. |
| [`make`](https://www.gnu.org/software/make/) | Run repository workflows. |

See [`docs/development.md`](docs/development.md) for the supported development environment and workflow-specific requirements.

Kubernetes validation is optional for contributors. It requires [`kubectl`](https://kubernetes.io/docs/reference/kubectl/), access to an operator-selected cluster, and the credentials described in `docs/operator-guide.md`.

## Development Workflow

1. Make the smallest change that solves the problem.
2. Add or update behavior-focused tests when runtime behavior changes.
3. Run the validation appropriate to the change before opening it for review.
4. Update the relevant source-of-truth document when a user-visible contract changes.

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

Keep guidance in its audience-specific guide and link to it instead of copying it. Keep the [README documentation table](README.md#documentation) current when adding or removing guides. The OpenAPI specification is canonical for HTTP paths, methods, statuses, schemas, and examples; use API Semantics for cross-cutting behavior that the contract cannot express clearly.

## Pull Request Expectations

A change should state its behavioral effect, affected interfaces, and verification performed. Keep unrelated cleanup out of the same change. Do not add telemetry, retries, compatibility aliases, or new abstractions without a concrete requirement.
