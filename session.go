package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// sanitizeContextName replaces characters that are unsafe or inconvenient in filenames.
func sanitizeContextName(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		" ", "_",
	)

	return replacer.Replace(name)
}

// createSessionFile creates a kubeconfig session file in the specified directory.
func createSessionFile(sessionsDir, contextName string, pid int, data []byte) (string, error) {
	if err := os.MkdirAll(sessionsDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create sessions directory: %w", err)
	}

	sanitized := sanitizeContextName(contextName)
	filename := fmt.Sprintf("%s.%d.yaml", sanitized, pid)
	sessionPath := filepath.Join(sessionsDir, filename)

	if err := os.WriteFile(sessionPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write session file: %w", err)
	}

	return sessionPath, nil
}

// isProcessAlive checks whether a process with given PID exists and is running.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds. Sending signal 0 performs error-checking
	// without actually sending a signal.
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	// If err is EPERM, the process exists and is running, but we lack permission to signal it.
	if errors.Is(err, syscall.EPERM) {
		return true
	}

	return false
}

// cleanupStaleSessions removes session files whose corresponding processes are no longer alive.
func cleanupStaleSessions(sessionsDir string) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}

		// Expected format: <context>.<pid>.yaml
		parts := strings.Split(name, ".")
		if len(parts) < 3 {
			continue
		}

		pidStr := parts[len(parts)-2]

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		if !isProcessAlive(pid) {
			_ = os.Remove(filepath.Join(sessionsDir, name))
		}
	}
}
