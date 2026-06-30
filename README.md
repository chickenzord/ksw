# ksw

Kubeconfig SWitcher

![Go version](https://img.shields.io/github/go-mod/go-version/chickenzord/ksw)
[![Go Report Card](https://goreportcard.com/badge/github.com/chickenzord/ksw)](https://goreportcard.com/report/github.com/chickenzord/ksw)
[![codecov](https://codecov.io/github/chickenzord/ksw/graph/badge.svg?token=VBb5SKLC4O)](https://codecov.io/github/chickenzord/ksw)
![GitHub release](https://img.shields.io/github/v/release/chickenzord/ksw)

`ksw` is a fast and lightweight Go CLI tool to switch Kubernetes contexts by starting a new shell with an isolated kubeconfig file. This allows different terminals to use different contexts concurrently without interfering with each other.

## Why ksw?

I wanted a kubeconfig switcher that is simple (as in Unix philosophy) and integrates easily with my existing ZSH and Prezto setup without getting in the way. It also needs to work with other Kubernetes tools without many changes.

Other solutions I tried:
- **kubectx / kubens**: Changing the context in one terminal changes it everywhere.
- **kubie**: It does too much and broke my ZSH setup.
- **kube_ps1**: I still use it to show the current context alongside `ksw`.

## Features

- **Isolated contexts per terminal**: Work with different Kubernetes contexts across multiple terminals simultaneously.
- **No nested shells**: Switching contexts in an active session updates the config file in-place instead of spawning new shells.
- **Fuzzy finder**: Shows a fuzzy finder to select a context if no exact match is specified.
- **Optional minification**: Can strip unused clusters, contexts, and users from the temporary kubeconfig.

## Installation

```sh
brew install chickenzord/tap/ksw
```

Alternatively, install from source:

```sh
go install github.com/chickenzord/ksw
```

## Configuration

`ksw` loads configuration from `~/.config/ksw/config.yaml` or `~/.ksw.yaml` (fallback).

```yaml
kubeconfig:
  # When true, extracts only the cluster, user, and context needed for the active context.
  # Defaults to false, which preserves other contexts but updates current-context.
  minify: false
```

## How it works

```sh
ksw [context-name]
```

### First time (not in a ksw session):
1. Loads kubeconfig from these locations (in order):
   - Path set in `KSW_KUBECONFIG_ORIGINAL`
   - Path set in `KUBECONFIG`
   - Default location `$HOME/.kube/config`
2. Evaluates configuration options. If `minify` is enabled, extracts only the cluster, user, and context for the specified context. Otherwise, copies the config and updates the `current-context`.
3. Writes the isolated config to a temporary file.
4. Replaces the `ksw` process with your shell using `syscall.Exec()`, setting `KUBECONFIG` to the temp file.
5. Your shell now uses the isolated context.

### When already in a ksw session:
1. Running `ksw [another-context]` detects you are already in a session.
2. Updates the existing temp kubeconfig file in-place.
3. Returns immediately. Kubectl sees the new context right away without spawning new shells.

### Environment variables set in the shell:
- `KSW_KUBECONFIG_ORIGINAL`: Path to your original kubeconfig file
- `KSW_KUBECONFIG`: Path to the temp isolated kubeconfig
- `KUBECONFIG`: Same as KSW_KUBECONFIG
- `KSW_ACTIVE`: Always set to "true" when in a ksw session
- `KSW_SHELL`: Path to your shell (e.g. `/bin/zsh`)

## Limitations

- No automatic prompt indicator. Use the environment variables (`KSW_ACTIVE`, `KSW_KUBECONFIG_ORIGINAL`) in your prompt setup.
- Temp kubeconfig files rely on OS cleanup (typically automatic in `/tmp`).
- Primarily tested on ZSH on Darwin Arm64.
