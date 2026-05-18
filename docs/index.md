---
layout: default
title: ksw - Kubeconfig SWitcher
description: A simple CLI tool that switches Kubernetes contexts by starting a new shell with an isolated, minified kubeconfig file
---

# ksw

**Kubeconfig SWitcher**

![Go version](https://img.shields.io/github/go-mod/go-version/chickenzord/ksw)
[![Go Report Card](https://goreportcard.com/badge/github.com/chickenzord/ksw)](https://goreportcard.com/report/github.com/chickenzord/ksw)
[![codecov](https://codecov.io/github/chickenzord/ksw/graph/badge.svg?token=VBb5SKLC4O)](https://codecov.io/github/chickenzord/ksw)
![GitHub release](https://img.shields.io/github/v/release/chickenzord/ksw)

---

## Overview

A simple CLI tool that switches Kubernetes contexts by starting a new shell with an isolated, minified kubeconfig file. Work with different Kubernetes contexts across multiple terminals simultaneously without conflicts.

## Installation

```sh
brew install chickenzord/tap/ksw
```

Alternatively, install from source:

```sh
go install github.com/chickenzord/ksw
```

## Quick Start

Switch to a specific context:

```sh
ksw context-name
```

Or use the built-in fuzzy finder to select a context:

```sh
ksw
```

## Features

- **Isolated contexts per terminal**: Work with different Kubernetes contexts across multiple terminals simultaneously without conflicts
- **No nested shells**: Switching contexts in a ksw session updates the config in-place
- **Process efficiency**: Uses `exec` to replace ksw process with your shell (v0.5.0+)
- **Fuzzy finder**: Built-in fuzzy finder (like [fzf](https://github.com/junegunn/fzf)) when no context specified
- **Minified configs**: Each session uses a minified kubeconfig with only the necessary context
- **Simple integration**: Works with any kubectl-compatible tool without configuration

## How It Works

### First time (not in a ksw session)

1. Loads kubeconfig from these locations (in order):
   - Path set in `KSW_KUBECONFIG_ORIGINAL`
   - Path set in `KUBECONFIG`
   - Default location `$HOME/.kube/config`
2. Minifies the config to only include the cluster, user, and context for the specified context
3. Writes the minified config to a temporary file
4. Replaces the ksw process with your shell using `exec`, setting `KUBECONFIG` to the temp file
5. Your shell now uses the isolated context - kubectl commands work immediately

### When already in a ksw session

1. Running `ksw another-context` detects you're already in a session
2. Updates the existing temp kubeconfig file with the new context
3. Returns immediately - kubectl sees the new context right away
4. No nested shells, no process spawning

> **Note:** Starting from v0.5.0, ksw uses `syscall.Exec()` to replace its process with your shell, eliminating the ksw process from the process tree. When switching contexts within a ksw session, the kubeconfig file is updated in-place without spawning new shells, avoiding any nesting issues.

## Environment Variables

When in a ksw session, these variables are available in your shell:

- `KSW_KUBECONFIG_ORIGINAL`: Path to your original kubeconfig file
- `KSW_KUBECONFIG`: Path to the temp minified kubeconfig
- `KUBECONFIG`: Same as KSW_KUBECONFIG (standard kubectl environment variable)
- `KSW_ACTIVE`: Always set to "true" when in a ksw session
- `KSW_SHELL`: Path to your shell (e.g. `/bin/zsh`)

## Why ksw?

I wanted a kubeconfig switcher that's simple (as in Unix philosophy) and can integrate easily with my existing ZSH and Prezto setup without getting in the way. It must also integrate with other Kubernetes tools without much configuration.

Other solutions I tried:

- **kubectx and kubens**: They are good, but I switch and use multiple contexts concurrently a lot. Changing context in one terminal will change other terminals as well because they share the same kubeconfig file.
- **kubie**: Took a lot of inspiration from this project. But somehow it's doing too much and messed with ZDOTDIR breaking my ZSH setup.
- **kube_ps1**: Still using this for showing current context, and it integrates well with ksw

## Limitations

- No automatic prompt indicator - use the environment variables (`KSW_ACTIVE`, `KSW_KUBECONFIG_ORIGINAL`) in your prompt setup
- Temp kubeconfig files rely on OS cleanup (typically automatic in `/tmp`)
- Primarily tested on ZSH on Darwin Arm64

## Links

- [GitHub Repository](https://github.com/chickenzord/ksw)
- [Latest Releases](https://github.com/chickenzord/ksw/releases)
- [Report Issues](https://github.com/chickenzord/ksw/issues)

---

Made with care for the Kubernetes community
