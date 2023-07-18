package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func currentLevel() int {
	if os.Getenv("KSW_LEVEL") == "" {
		return 0
	}

	level, err := strconv.Atoi(os.Getenv("KSW_LEVEL"))
	if err != nil {
		panic("invalid KSW_LEVEL")
	}

	return level
}

func startShell(shell, contextName string) error {
	var kubeconfigOriginal string // TODO add condition to use original kubeconfig from cli flags
	if path := os.Getenv("KSW_KUBECONFIG_ORIGINAL"); path != "" {
		kubeconfigOriginal = path
	} else if path := os.Getenv("KUBECONFIG"); path != "" {
		kubeconfigOriginal = path
	} else {
		kubeconfigOriginal = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	b, err := generateKubeconfig(kubeconfigOriginal, contextName)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp("", fmt.Sprintf("%s.*.yaml", contextName))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		return err
	}

	os.Setenv("KUBECONFIG", f.Name())
	os.Setenv("KSW_KUBECONFIG_ORIGINAL", kubeconfigOriginal)
	os.Setenv("KSW_KUBECONFIG", f.Name())
	os.Setenv("KSW_ACTIVE", "true")
	os.Setenv("KSW_SHELL", shell)
	os.Setenv("KSW_LEVEL", fmt.Sprintf("%d", currentLevel()+1))
	os.Setenv("KSW_CONTEXT", contextName)

	fmt.Printf("Starting shell for context %s\n", contextName)
	defer func(contextName string) {
		fmt.Printf("Exited from context %s\n", contextName)
	}(contextName)

	sh := exec.Command(shell)
	sh.Stderr = os.Stderr
	sh.Stdin = os.Stdin
	sh.Stdout = os.Stdout

	if err := sh.Run(); err != nil {
		return err
	}

	return nil
}
