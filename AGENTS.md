# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ksw** (Kubeconfig SWitcher) is a CLI tool written in Go that switches Kubernetes contexts by starting a new shell with an isolated, minified kubeconfig file. This allows multiple shells to use different contexts concurrently without interfering with each other.

## Architecture

### Core Workflow

1. **Kubeconfig Loading** (kubeconfig.go:48-59): Loads kubeconfig from `KSW_KUBECONFIG_ORIGINAL`, `KUBECONFIG`, or `$HOME/.kube/config`
2. **Minification** (kubeconfig.go:14-46): Extracts only the cluster, user, and context needed for the specified context
3. **Shell Execution** (shell.go:20-63): Creates a temporary kubeconfig file and uses `syscall.Exec()` to replace ksw process with shell
4. **Context Switching** (shell.go:73-98): When already in a ksw session, updates kubeconfig file in-place without spawning new processes
5. **Context Selection** (kubeconfig.go:107-138): Tries exact match first, then shows fuzzy finder UI if no match found
6. **Cleanup**: Temporary kubeconfig files rely on OS temp directory cleanup (no explicit cleanup mechanism)

### Key Components

- **main.go**: Entry point using urfave/cli framework; detects if already in ksw session via `KSW_KUBECONFIG_ORIGINAL`; supports `--list` flag to list contexts
- **kubeconfig.go**: Kubeconfig parsing, minification, and context selection using k8s.io/client-go API types (apiv1.Config)
- **shell.go**: Shell execution using `syscall.Exec()` and temp file management
- **version.go**: Version information structure with build-time variables (Version, GitCommit, BuildDate)
- **log.go**: Simple logging utility

### Environment Variables Set by ksw

When starting a shell, ksw sets these variables:
- `KSW_KUBECONFIG_ORIGINAL`: Original kubeconfig path (used to detect existing ksw sessions)
- `KSW_KUBECONFIG`: Path to the temporary minified kubeconfig
- `KUBECONFIG`: Same as KSW_KUBECONFIG (standard kubectl variable)
- `KSW_ACTIVE`: Always "true" when inside a ksw shell
- `KSW_SHELL`: Path to the shell executable

## Development Commands

### Building
```bash
go build
```

### Testing
```bash
# Run all tests with race detection and coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests for a specific file
go test -v ./kubeconfig_test.go
```

### Linting
```bash
golangci-lint run
```

The project uses golangci-lint with these enabled linters: errcheck, govet, ineffassign, staticcheck, unused, gocyclo, revive, whitespace, wsl_v5. The `exported` rule in revive is disabled since this is not a library project.

### Installing Locally
```bash
go install
```

### Release
Releases are automated using goreleaser. The build:
- Targets Linux, Windows, and Darwin (amd64 and arm64)
- Uses CGO_ENABLED=0 for static binaries
- Publishes to chickenzord/homebrew-tap
- Sets version info via ldflags: `-X main.Version={{.Version}} -X main.GitCommit={{.FullCommit}} -X main.BuildDate={{.Date}}`
- Commit messages must follow conventional commits format (enforced by CI)

## Dependencies

- **github.com/urfave/cli/v2**: CLI framework
- **github.com/riywo/loginshell**: Detects user's login shell
- **github.com/ktr0731/go-fuzzyfinder**: Interactive fuzzy finder
- **k8s.io/client-go**: Kubernetes client-go for kubeconfig types (apiv1.Config)
- **github.com/ghodss/yaml**: YAML marshaling/unmarshaling

## Code Patterns

- All files are in the `main` package (single-package binary)
- Version variables (Version, GitCommit, BuildDate) are set via ldflags at build time (version.go:9-14)
- Uses `syscall.Exec()` to replace ksw process with shell (shell.go:58), so ksw never appears in process tree
- Temporary kubeconfig files created with pattern `{context-name}.*.yaml` in OS temp directory
- Temp file cleanup relies on OS temp directory cleanup (no explicit cleanup mechanism)
- When already in a ksw session (`KSW_KUBECONFIG_ORIGINAL` is set), updates temp file in-place instead of creating new session (main.go:55-56)
- Context selection tries exact match first before showing fuzzy finder (kubeconfig.go:117-122)
- Contexts are sorted alphabetically before being shown in fuzzy finder (kubeconfig.go:115)
