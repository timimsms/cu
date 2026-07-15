# Contributing to cu

Thanks for your interest in contributing! `cu` is an unofficial, community-maintained ClickUp CLI, and contributions of all kinds are welcome.

## Development Setup

Prerequisites:

- Go 1.24 or later
- `make` (optional but convenient)

Clone and build:

```bash
git clone https://github.com/timimsms/cu.git
cd cu
make build     # builds the ./cu binary
```

Common tasks:

```bash
make build     # build the binary
make test      # run tests
make lint      # run golangci-lint (brew install golangci-lint)
make fmt       # gofmt + go mod tidy
```

Before pushing, run the full local CI suite (mirrors GitHub Actions — vet, staticcheck, gosec, errcheck, tests with race detector, build, tidy check, formatting):

```bash
./scripts/ci.sh    # or: make ci
```

## Commit Messages

This repo uses conventional-commit prefixes:

- `feat:` — new features
- `fix:` — bug fixes
- `docs:` — documentation changes
- `chore:` — maintenance, tooling, dependencies
- `ci:` — CI workflow changes
- `test:` — adding or improving tests

## Testing

- All changes should pass `go test -race ./...`.
- Add tests for new behavior where practical; table-driven tests are preferred.

## Regenerating CLI Docs

The command reference under `docs/site/commands/` is generated from the cobra command definitions. If you change command help text, regenerate the docs:

```bash
make build
./cu docs markdown --dir docs/site/commands
```

Commit the regenerated files along with your change.

## Previewing the Documentation Site

The docs site uses [MkDocs](https://www.mkdocs.org/) with the Material theme:

```bash
pip install mkdocs-material
mkdocs serve    # local preview at http://127.0.0.1:8000
```

`mkdocs build --strict` must pass — CI treats warnings as errors.

## Pull Requests

- Branch from `main` and keep PRs focused on a single change.
- Fill out the PR template and make sure CI is green.
