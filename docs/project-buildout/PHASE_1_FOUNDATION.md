# Phase 1: Foundation & Setup

## Overview
Establish the core project structure, development environment, and foundational components for the ClickUp CLI.

## Prerequisites
- [x] Go 1.22+ installed
- [x] Git configured
- [x] GitHub repository access
- [x] ClickUp account with API access

## Task Checklist

### 1. Project Initialization
- [x] Initialize Go module: `go mod init github.com/tim/cu`
- [x] Create standard Go project structure:
  ```
  ├── cmd/
  │   └── cu/
  │       └── main.go
  ├── internal/
  │   ├── api/
  │   ├── auth/
  │   ├── cache/
  │   ├── cmd/
  │   ├── config/
  │   ├── errors/
  │   ├── output/
  │   └── version/
  ├── scripts/
  ├── test/
  └── .github/
      └── workflows/
  ```
- [x] Set up `.gitignore` for Go projects
- [x] Create initial `README.md` with project overview
- [x] Add MIT LICENSE file

### 2. Cobra CLI Framework Setup
- [x] Add Cobra dependency: `go get github.com/spf13/cobra@latest`
- [x] Add Viper for configuration: `go get github.com/spf13/viper@latest`
- [x] Create root command structure in `cmd/cu/main.go`
- [x] Implement basic command scaffolding:
  - [x] `cu --version`
  - [x] `cu --help`
  - [x] `cu completion` (bash, zsh, fish, powershell)
- [x] Set up command hierarchy with placeholders

### 3. Configuration Management
- [x] Design configuration schema (YAML structure)
- [x] Implement config file locations:
  - [x] User config: `~/.config/cu/config.yml`
  - [ ] Project config: `.cu.yml`
  - [x] Environment variables: `CU_*`
- [x] Create config command structure:
  - [x] `cu config list`
  - [x] `cu config get <key>`
  - [x] `cu config set <key> <value>`
- [x] Implement config precedence (env > project > user > defaults)
- [x] Add config validation

### 4. Authentication Framework
- [x] Design auth token storage using OS keychain:
  - [x] macOS: Keychain Access
  - [x] Linux: Secret Service API
  - [x] Windows: Credential Manager
- [x] Implement `cu auth` command structure:
  - [x] `cu auth login` (interactive)
  - [x] `cu auth login --token <token>`
  - [x] `cu auth status`
  - [x] `cu auth logout`
- [x] Create auth token validation
- [x] Add multi-workspace support (token per workspace)
- [x] Implement secure token storage/retrieval

### 5. API Client Foundation
- [x] Evaluate and integrate go-clickup SDK
- [x] Create API client wrapper in `internal/api/`
- [x] Implement:
  - [x] Request authentication
  - [x] Rate limiting (100 req/min for free tier)
  - [x] Exponential backoff for 429 errors
  - [ ] Request/response logging (debug mode)
- [x] Add API error handling and user-friendly messages
- [ ] Create mock client for testing

### 6. Output Formatting System
- [x] Create output formatter interface in `internal/output/`
- [x] Implement formatters:
  - [x] Table (human-readable, default)
  - [x] JSON (machine-readable)
  - [x] YAML
  - [x] CSV/TSV
- [x] Add output flags to root command:
  - [x] `--output` / `-o` (json|yaml|table|csv)
  - [ ] `--filter` / `-f` (jq-style filtering)
- [ ] Implement column selection for table output
- [x] Add color support with disable flag

### 7. Error Handling & Logging
- [x] Create consistent error types in `internal/errors/`
- [x] Implement error wrapping with context
- [x] Add debug logging with `--debug` flag
- [x] Create user-friendly error messages for:
  - [x] Network failures
  - [x] Authentication errors
  - [x] Rate limiting
  - [x] Invalid input
  - [x] API errors
- [ ] Implement error reporting for crashes (opt-in)

### 8. Testing Infrastructure
- [x] Set up Go testing structure
- [x] Add testify for assertions: `go get github.com/stretchr/testify`
- [ ] Create test fixtures for API responses
- [ ] Implement test helpers in `test/`
- [x] Add unit tests for:
  - [x] Configuration management
  - [ ] Authentication
  - [ ] Output formatting
  - [ ] Error handling
- [ ] Create integration test framework
- [x] Add code coverage reporting

### 9. Development Tooling
- [x] Set up script with common tasks:
  - [x] `scripts/ci.sh` for local CI
  - [x] Build, test, lint, format checks
  - [x] Security scanning with gosec
  - [x] Error checking with errcheck
- [x] Configure staticcheck
- [ ] Add pre-commit hooks:
  - [ ] gofmt
  - [ ] golangci-lint
  - [ ] go test
- [x] Set up EditorConfig
- [ ] Create development documentation

### 10. CI/CD Pipeline (GitHub Actions)
- [x] Create `.github/workflows/ci.yml`:
  - [x] Run on PR and main branch
  - [x] Matrix build (Go versions)
  - [x] Run tests with coverage
  - [x] Run linters
  - [x] Security scanning (gosec)
- [ ] Create `.github/workflows/release.yml`:
  - [ ] Trigger on version tags
  - [ ] Build binaries for all platforms
  - [ ] Create GitHub release
  - [ ] Upload artifacts
- [x] Add status badges to README

### 11. Documentation Foundation
- [ ] Create `docs/` structure:
  ```
  docs/
  ├── installation.md
  ├── configuration.md
  ├── authentication.md
  ├── commands/
  └── development.md
  ```
- [ ] Write initial documentation:
  - [ ] Installation guide (placeholder)
  - [ ] Configuration reference
  - [ ] Authentication setup
  - [ ] Development setup
- [ ] Add inline code documentation
- [ ] Create CONTRIBUTING.md

### 12. Version Management
- [x] Implement version package in `internal/version/`
- [x] Add build-time version injection
- [x] Create version command with:
  - [x] Version number
  - [x] Build date
  - [x] Git commit hash
  - [x] Go version
- [ ] Add update check mechanism (opt-in)
- [ ] Implement version comparison logic

## Validation Checklist
- [x] `go build ./cmd/cu` produces binary
- [x] `cu --version` shows version info
- [x] `cu --help` displays command tree
- [x] `cu auth login --token <token>` stores token securely
- [x] `cu config set default_space "MySpace"` persists config
- [x] All tests pass: `go test ./...`
- [x] Linting passes: `staticcheck ./...`
- [x] CI pipeline runs successfully

## Next Steps
Once Phase 1 is complete, proceed to [Phase 2: Core Features](./PHASE_2_CORE_FEATURES.md) to implement the main ClickUp operations.