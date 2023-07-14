# ksw

Kubeconfig SWitcher

[![Go Report Card](https://goreportcard.com/badge/github.com/chickenzord/ksw)](https://goreportcard.com/report/github.com/chickenzord/ksw)


## Installation

```sh
brew install chickenzord/tap/ksw
```

Alternatively you can also install from the source by running `go install github.com/chickenzord/ksw` (requires Go build tools)

## How it works

```sh
ksw context-name
```

1. Try loading kubeconfig file from these locations:
   1. Path set in `KSW_KUBECONFIG_ORIGINAL` (more on this below)
   2. Path set in `KUBECONFIG`
   3. Default location `$HOME/.kube/config`
2. Minify and flatten the config so it only contains clusters and users used by the specificed "context-name", then put it in a temp file
3. Start a new shell ([same with the currently used](https://github.com/riywo/loginshell)) with `KUBECONFIG` set to the temp file
4. Additionally, these environment variables also set in the sub-shell:
   - `KSW_KUBECONFIG_ORIGINAL`: To keep track of original kubeconfig file when starting recursive shells
   - `KSW_KUBECONFIG`: Same value as KUBECONFIG
   - `KSW_ACTIVE`: Always set to "true"
   - `KSW_SHELL`: Path to the shell (e.g. `/bin/zsh`)
   - `KSW_LEVEL`: Nesting level of the shell, starting at 1 when first running ksw
   - `KSW_CONTEXT`: Kube context name used when running ksw

## Wait, why am I creating this?

I want a kubeconfig switcher that simple (as in Unix philosophy) and can integrate easily with my existing ZSH and Prezto setup without getting in the way. Must also be able to integrate with other kubernetes tools without much changes.

Other solutions I have tried:

- kubectx and kubens: They are good, but I switch and use multiple contexts concurrently a lot. Changing context in one terminal will change other terminals as well because they are sharing the same kubeconfig file.
- kubie: Took a lot of inspirations from this project. But somehow it's doing too much and messed with ZDOTDIR breaking my ZSH setup.
- kube_ps1: Still using this for showing current context, and it integrates well with ksw
