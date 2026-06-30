# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ksw** (Kubeconfig SWitcher) is a CLI tool written in Go that switches Kubernetes contexts by starting a new shell with an isolated, minified kubeconfig file. This allows multiple shells to use different contexts concurrently without interfering with each other.

## Architecture & Design

### Core Workflow

1. **Kubeconfig Loading**: Loads kubeconfig from `KSW_KUBECONFIG_ORIGINAL`, `KUBECONFIG`, or `$HOME/.kube/config`.
2. **Minification**: Extracts only the cluster, user, and context needed for the specified context.
3. **Shell Execution**: Creates a temporary kubeconfig file and uses `syscall.Exec()` to replace the ksw process with the shell.
4. **Context Switching**: When already in a ksw session, updates the kubeconfig file in-place without spawning new processes.
5. **Context Selection**: Tries exact match first, then shows a fuzzy finder UI if no match is found.
6. **Cleanup**: Temporary kubeconfig files rely on OS temp directory cleanup (no explicit cleanup mechanism).

### Environment Variables

When starting a shell, ksw sets these variables:
- `KSW_KUBECONFIG_ORIGINAL`: Original kubeconfig path (used to detect existing ksw sessions).
- `KSW_KUBECONFIG`: Path to the temporary minified kubeconfig.
- `KUBECONFIG`: Same as KSW_KUBECONFIG (standard kubectl variable).
- `KSW_ACTIVE`: Always "true" when inside a ksw shell.
- `KSW_SHELL`: Path to the shell executable.

## Git Commit Conventions

- Commit messages must follow the Conventional Commits format (enforced by CI).
- Conventional git scope must never exceed 1 (either 0 or 1).

## Code Patterns

- All files are in the `main` package (single-package binary).
- Version variables (`Version`, `GitCommit`, `BuildDate`) are set via ldflags at build time.
- Uses `syscall.Exec()` to replace the ksw process with a shell, so ksw never appears in the process tree.
- Temporary kubeconfig files are created with the pattern `{context-name}.*.yaml` in the OS temp directory.
- Temp file cleanup relies on OS temp directory cleanup (no explicit cleanup mechanism).
- When already in a ksw session (`KSW_KUBECONFIG_ORIGINAL` is set), updates the temp file in-place instead of creating a new session.
- Context selection tries an exact match first before showing the fuzzy finder.
- Contexts are sorted alphabetically before being shown in the fuzzy finder.
