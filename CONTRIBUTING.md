# Contributing

## Prerequisites

Install or enable the following before contributing:

| Requirement | Used for |
| --- | --- |
| [Git](https://git-scm.com/) | Clone the repository and manage changes. |
| [Go](https://go.dev/dl/) | Build, test, and run Anya locally. |
| [Docker Engine](https://docs.docker.com/engine/) with [Buildx](https://docs.docker.com/reference/cli/docker/buildx/) support | Build and smoke-test the container image. |
| [`make`](https://www.gnu.org/software/make/) | Run repository workflows. |

See [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for the supported development environment and workflow-specific requirements.

Kubernetes validation is optional for contributors. It requires [`kubectl`](https://kubernetes.io/docs/reference/kubectl/), access to an operator-selected cluster, and the credentials described in `docs/OPERATOR-GUIDE.md`.

## Development Workflow

1. Read [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for the supported local workflow and environment requirements.
2. Make the smallest change that solves the problem.
3. Add or update behavior-focused tests when runtime behavior changes.
4. Run the validation appropriate to the change, then run `make ci` before opening a change for review.
5. Update the relevant source-of-truth document when a user-visible contract changes.

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
- `docs/OPERATOR-GUIDE.md`: deployment and operator procedures.
- `CONTRIBUTING.md`: repository contribution workflow.

## Pull Request Expectations

A change should state its behavioral effect, affected interfaces, and verification performed. Keep unrelated cleanup out of the same change. Do not add telemetry, retries, compatibility aliases, or new abstractions without a concrete requirement.
