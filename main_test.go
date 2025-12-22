package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestEnvAction(t *testing.T) {
	// Save original env vars
	origKswActive := os.Getenv("KSW_ACTIVE")
	origKswKubeconfig := os.Getenv("KSW_KUBECONFIG")
	origKswKubeconfigOriginal := os.Getenv("KSW_KUBECONFIG_ORIGINAL")
	origKswShell := os.Getenv("KSW_SHELL")
	origKubeconfig := os.Getenv("KUBECONFIG")

	// Restore after test
	defer func() {
		_ = os.Setenv("KSW_ACTIVE", origKswActive)
		_ = os.Setenv("KSW_KUBECONFIG", origKswKubeconfig)
		_ = os.Setenv("KSW_KUBECONFIG_ORIGINAL", origKswKubeconfigOriginal)
		_ = os.Setenv("KSW_SHELL", origKswShell)
		_ = os.Setenv("KUBECONFIG", origKubeconfig)
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		wantVars []string
	}{
		{
			name: "all env vars set",
			envVars: map[string]string{
				"KSW_ACTIVE":              "true",
				"KSW_KUBECONFIG":          "/tmp/test.yaml",
				"KSW_KUBECONFIG_ORIGINAL": "/home/user/.kube/config",
				"KSW_SHELL":               "/bin/zsh",
				"KUBECONFIG":              "/tmp/test.yaml",
			},
			wantVars: []string{
				"KSW_ACTIVE=true",
				"KSW_KUBECONFIG=/tmp/test.yaml",
				"KSW_KUBECONFIG_ORIGINAL=/home/user/.kube/config",
				"KSW_SHELL=/bin/zsh",
				"KUBECONFIG=/tmp/test.yaml",
			},
		},
		{
			name: "some env vars empty",
			envVars: map[string]string{
				"KSW_ACTIVE":              "",
				"KSW_KUBECONFIG":          "",
				"KSW_KUBECONFIG_ORIGINAL": "/home/user/.kube/config",
				"KSW_SHELL":               "/bin/bash",
				"KUBECONFIG":              "/home/user/.kube/config",
			},
			wantVars: []string{
				"KSW_ACTIVE=",
				"KSW_KUBECONFIG=",
				"KSW_KUBECONFIG_ORIGINAL=/home/user/.kube/config",
				"KSW_SHELL=/bin/bash",
				"KUBECONFIG=/home/user/.kube/config",
			},
		},
		{
			name: "all env vars empty",
			envVars: map[string]string{
				"KSW_ACTIVE":              "",
				"KSW_KUBECONFIG":          "",
				"KSW_KUBECONFIG_ORIGINAL": "",
				"KSW_SHELL":               "",
				"KUBECONFIG":              "",
			},
			wantVars: []string{
				"KSW_ACTIVE=",
				"KSW_KUBECONFIG=",
				"KSW_KUBECONFIG_ORIGINAL=",
				"KSW_SHELL=",
				"KUBECONFIG=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			_ = os.Unsetenv("KSW_ACTIVE")
			_ = os.Unsetenv("KSW_KUBECONFIG")
			_ = os.Unsetenv("KSW_KUBECONFIG_ORIGINAL")
			_ = os.Unsetenv("KSW_SHELL")
			_ = os.Unsetenv("KUBECONFIG")

			// Set env vars for this test case
			for key, value := range tt.envVars {
				if value != "" {
					_ = os.Setenv(key, value)
				}
			}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run envAction
			err := envAction()
			if err != nil {
				t.Errorf("envAction() returned error: %v", err)
			}

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Verify all expected env vars are in output
			for _, wantVar := range tt.wantVars {
				if !strings.Contains(output, wantVar) {
					t.Errorf("envAction() output missing %q\nGot output:\n%s", wantVar, output)
				}
			}

			// Verify output has exactly 5 lines (one for each env var)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) != 5 {
				t.Errorf("envAction() output has %d lines, want 5\nGot output:\n%s", len(lines), output)
			}
		})
	}
}

func TestEnvActionNoError(t *testing.T) {
	// Simple test to ensure envAction returns no error
	err := envAction()
	if err != nil {
		t.Errorf("envAction() returned unexpected error: %v", err)
	}
}
