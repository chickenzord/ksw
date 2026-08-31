package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeContextName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"minikube", "minikube"},
		{"arn:aws:eks:us-west-2:123456789012:cluster/my-cluster", "arn-aws-eks-us-west-2-123456789012-cluster-my-cluster"},
		{"gke_project_zone_cluster", "gke_project_zone_cluster"},
		{"cluster\\name/test:1", "cluster-name-test-1"},
		{"my context", "my_context"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeContextName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeContextName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateSessionFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	contextName := "arn:aws:eks:cluster/demo"
	pid := 12345
	data := []byte("apiVersion: v1\nkind: Config\n")

	filePath, err := createSessionFile(sessionsDir, contextName, pid, data)
	if err != nil {
		t.Fatalf("createSessionFile() failed: %v", err)
	}

	expectedFileName := fmt.Sprintf("%s.%d.yaml", sanitizeContextName(contextName), pid)

	expectedPath := filepath.Join(sessionsDir, expectedFileName)
	if filePath != expectedPath {
		t.Errorf("filePath = %q, want %q", filePath, expectedPath)
	}

	// Verify file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("content = %q, want %q", string(content), string(data))
	}

	// Verify file permissions (0600)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("os.Stat() failed: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %v, want 0600", info.Mode().Perm())
	}
}

func TestCleanupStaleSessions(t *testing.T) {
	sessionsDir := t.TempDir()

	// 1. Session file belonging to current process (alive)
	alivePid := os.Getpid()

	aliveFile := filepath.Join(sessionsDir, fmt.Sprintf("alive-context.%d.yaml", alivePid))
	if err := os.WriteFile(aliveFile, []byte("alive"), 0600); err != nil {
		t.Fatalf("failed to write alive file: %v", err)
	}

	// 2. Session file belonging to a non-existent PID (dead)
	// PID 9999999 is very unlikely to be active
	deadPid := 9999999

	deadFile := filepath.Join(sessionsDir, fmt.Sprintf("dead-context.%d.yaml", deadPid))
	if err := os.WriteFile(deadFile, []byte("dead"), 0600); err != nil {
		t.Fatalf("failed to write dead file: %v", err)
	}

	// 3. Unrelated file that should not be touched
	otherFile := filepath.Join(sessionsDir, "README.txt")
	if err := os.WriteFile(otherFile, []byte("info"), 0644); err != nil {
		t.Fatalf("failed to write other file: %v", err)
	}

	cleanupStaleSessions(sessionsDir)

	// Alive file should still exist
	if _, err := os.Stat(aliveFile); os.IsNotExist(err) {
		t.Errorf("alive file was unexpectedly deleted")
	}

	// Dead file should be deleted
	if _, err := os.Stat(deadFile); !os.IsNotExist(err) {
		t.Errorf("dead file was not cleaned up")
	}

	// Other file should remain untouched
	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
		t.Errorf("unrelated file was deleted")
	}
}

func TestCreateSessionFile_NestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Test multi-level nested directory creation
	sessionsDir := filepath.Join(tmpDir, "deep", "nested", "sessions", "dir")
	contextName := "nested-ctx"
	pid := os.Getpid()
	data := []byte("apiVersion: v1\nkind: Config\n")

	filePath, err := createSessionFile(sessionsDir, contextName, pid, data)
	if err != nil {
		t.Fatalf("createSessionFile() failed to create nested directory: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("expected session file to exist at %s", filePath)
	}
}

func TestIsProcessAlive(t *testing.T) {
	// Current process must be alive
	if !isProcessAlive(os.Getpid()) {
		t.Errorf("expected current process PID %d to be alive", os.Getpid())
	}

	// PID 1 (init / launchd) is always alive
	if !isProcessAlive(1) {
		t.Errorf("expected PID 1 (init/launchd) to be alive")
	}

	// Non-positive PIDs should return false
	if isProcessAlive(0) {
		t.Errorf("expected PID 0 to not be reported alive")
	}

	if isProcessAlive(-1) {
		t.Errorf("expected PID -1 to not be reported alive")
	}

	// Non-existent PID should return false
	if isProcessAlive(9999999) {
		t.Errorf("expected non-existent PID 9999999 to not be alive")
	}
}

