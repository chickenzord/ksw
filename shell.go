package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// startShell creates a new ksw session by generating a minified kubeconfig
// and replacing the current process with the user's shell using syscall.Exec.
//
// It loads the original kubeconfig from KSW_KUBECONFIG_ORIGINAL, KUBECONFIG,
// or $HOME/.kube/config (in that order), minifies it to include only the
// specified context, writes it to a temporary file, and sets up environment
// variables before executing the shell.
//
// The ksw process is replaced entirely, so this function never returns on success.
// Temporary kubeconfig files are cleaned up by the OS temp directory cleanup.
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

	defer func() {
		_ = f.Close()
	}()

	if _, err := f.Write(b); err != nil {
		return err
	}

	_ = os.Setenv("KUBECONFIG", f.Name())
	_ = os.Setenv("KSW_KUBECONFIG_ORIGINAL", kubeconfigOriginal)
	_ = os.Setenv("KSW_KUBECONFIG", f.Name())
	_ = os.Setenv("KSW_ACTIVE", "true")
	_ = os.Setenv("KSW_SHELL", shell)

	logf("starting shell for context %s", contextName)

	// Replace ksw process with shell
	// Temp file cleanup relies on OS temp directory cleanup
	if err := syscall.Exec(shell, []string{shell}, os.Environ()); err != nil {
		return fmt.Errorf("failed to exec shell: %w", err)
	}

	return nil
}

// switchContext updates an existing ksw session to use a different Kubernetes context.
//
// This function is called when already inside a ksw session (KSW_KUBECONFIG_ORIGINAL is set).
// It regenerates the minified kubeconfig for the new context and overwrites the existing
// temporary kubeconfig file in-place. kubectl and other tools will immediately see the
// new context without requiring a new shell or process.
//
// This approach avoids nested shells and keeps the same process tree level.
func switchContext(contextName string) error {
	kubeconfigOriginal := os.Getenv("KSW_KUBECONFIG_ORIGINAL")
	if kubeconfigOriginal == "" {
		return fmt.Errorf("KSW_KUBECONFIG_ORIGINAL not set, cannot switch context")
	}

	existingKubeconfig := os.Getenv("KSW_KUBECONFIG")
	if existingKubeconfig == "" {
		return fmt.Errorf("KSW_KUBECONFIG not set, cannot switch context")
	}

	b, err := generateKubeconfig(kubeconfigOriginal, contextName)
	if err != nil {
		return err
	}

	// Overwrite existing temp file with new context
	if err := os.WriteFile(existingKubeconfig, b, 0600); err != nil {
		return err
	}

	logf("switched to context %s", contextName)

	// No process spawning - kubectl will immediately see the new context
	return nil
}
