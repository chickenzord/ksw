package main

import (
	"github.com/ktr0731/go-fuzzyfinder"
)

func pickContext() (string, error) {
	contexts, err := listContexts(getOriginalKubeconfigPath())
	if err != nil {
		return "", err
	}

	i, err := fuzzyfinder.Find(contexts, func(i int) string { return contexts[i] })
	if err != nil {
		return "", err
	}

	return contexts[i], nil
}
