package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestLoadConfig(t *testing.T) {
	origUserHomeDir := userHomeDir

	defer func() {
		userHomeDir = origUserHomeDir
	}()

	tests := []struct {
		name            string
		homeSetup       func(t *testing.T, home string)
		homeErr         error
		wantMinify      bool
		wantMergeOnExit bool
	}{
		{
			name:            "no config files - defaults to minify: false, merge_on_exit: false",
			homeSetup:       func(t *testing.T, home string) {},
			wantMinify:      false,
			wantMergeOnExit: false,
		},
		{
			name: "primary config file with minify: true",
			homeSetup: func(t *testing.T, home string) {
				configDir := filepath.Join(home, ".config", "ksw")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}

				content := []byte("kubeconfig:\n  minify: true\n")
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      true,
			wantMergeOnExit: false,
		},
		{
			name: "primary config file with minify: false",
			homeSetup: func(t *testing.T, home string) {
				configDir := filepath.Join(home, ".config", "ksw")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}

				content := []byte("kubeconfig:\n  minify: false\n")
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      false,
			wantMergeOnExit: false,
		},
		{
			name: "fallback config file with minify: true",
			homeSetup: func(t *testing.T, home string) {
				content := []byte("kubeconfig:\n  minify: true\n")
				if err := os.WriteFile(filepath.Join(home, ".ksw.yaml"), content, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      true,
			wantMergeOnExit: false,
		},
		{
			name: "both files exist - primary takes precedence",
			homeSetup: func(t *testing.T, home string) {
				configDir := filepath.Join(home, ".config", "ksw")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}

				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("kubeconfig:\n  minify: true\n"), 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}

				if err := os.WriteFile(filepath.Join(home, ".ksw.yaml"), []byte("kubeconfig:\n  minify: false\n"), 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      true,
			wantMergeOnExit: false,
		},
		{
			name: "invalid YAML - defaults to minify: false, merge_on_exit: false",
			homeSetup: func(t *testing.T, home string) {
				content := []byte("kubeconfig: {\ninvalid yaml")
				if err := os.WriteFile(filepath.Join(home, ".ksw.yaml"), content, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      false,
			wantMergeOnExit: false,
		},
		{
			name:            "home dir lookup error - defaults to minify: false, merge_on_exit: false",
			homeErr:         errors.New("home error"),
			wantMinify:      false,
			wantMergeOnExit: false,
		},
		{
			name: "primary config with merge_on_exit enabled",
			homeSetup: func(t *testing.T, home string) {
				configDir := filepath.Join(home, ".config", "ksw")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}

				content := []byte("kubeconfig:\n  merge_on_exit:\n    enabled: true\n")
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      false,
			wantMergeOnExit: true,
		},
		{
			name: "fallback config with merge_on_exit enabled",
			homeSetup: func(t *testing.T, home string) {
				content := []byte("kubeconfig:\n  merge_on_exit:\n    enabled: true\n")
				if err := os.WriteFile(filepath.Join(home, ".ksw.yaml"), content, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			wantMinify:      false,
			wantMergeOnExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if tt.homeSetup != nil {
				tt.homeSetup(t, tmpDir)
			}

			userHomeDir = func() (string, error) {
				if tt.homeErr != nil {
					return "", tt.homeErr
				}

				return tmpDir, nil
			}

			cfg := loadConfig()
			if cfg.Kubeconfig.Minify != tt.wantMinify {
				t.Errorf("loadConfig() Minify = %v, want %v", cfg.Kubeconfig.Minify, tt.wantMinify)
			}

			if cfg.Kubeconfig.MergeOnExit.Enabled != tt.wantMergeOnExit {
				t.Errorf("loadConfig() MergeOnExit.Enabled = %v, want %v", cfg.Kubeconfig.MergeOnExit.Enabled, tt.wantMergeOnExit)
			}
		})
	}
}

func TestGenerateKubeconfig_MinifyToggle(t *testing.T) {
	origUserHomeDir := userHomeDir

	defer func() {
		userHomeDir = origUserHomeDir
	}()

	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "source-kubeconfig")

	kubeconfigContent := `apiVersion: v1
kind: Config
current-context: dev-cluster
contexts:
- name: prod-cluster
  context:
    cluster: prod
    user: prod-user
- name: dev-cluster
  context:
    cluster: dev
    user: dev-user
clusters:
- name: prod
  cluster:
    server: https://prod.example.com
- name: dev
  cluster:
    server: https://dev.example.com
users:
- name: prod-user
  user:
    token: prod-token
- name: dev-user
  user:
    token: dev-token
`

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0600); err != nil {
		t.Fatalf("Failed to create test kubeconfig: %v", err)
	}

	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(filepath.Join(homeDir, ".config", "ksw"), 0755); err != nil {
		t.Fatalf("Failed to create home config directory: %v", err)
	}

	userHomeDir = func() (string, error) {
		return homeDir, nil
	}

	t.Run("minify is true", func(t *testing.T) {
		configContent := []byte("kubeconfig:\n  minify: true\n")
		if err := os.WriteFile(filepath.Join(homeDir, ".config", "ksw", "config.yaml"), configContent, 0600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		gotBytes, err := generateKubeconfig(kubeconfigPath, "prod-cluster")
		if err != nil {
			t.Fatalf("generateKubeconfig() error = %v", err)
		}

		var gotConfig apiv1.Config
		if err := yaml.Unmarshal(gotBytes, &gotConfig); err != nil {
			t.Fatalf("Failed to unmarshal generated kubeconfig: %v", err)
		}

		if gotConfig.CurrentContext != "prod-cluster" {
			t.Errorf("expected CurrentContext 'prod-cluster', got '%s'", gotConfig.CurrentContext)
		}

		if len(gotConfig.Contexts) != 1 || gotConfig.Contexts[0].Name != "prod-cluster" {
			t.Errorf("expected exactly 1 context 'prod-cluster', got context count: %d", len(gotConfig.Contexts))
		}

		if len(gotConfig.Clusters) != 1 || gotConfig.Clusters[0].Name != "prod" {
			t.Errorf("expected exactly 1 cluster 'prod', got cluster count: %d", len(gotConfig.Clusters))
		}

		if len(gotConfig.AuthInfos) != 1 || gotConfig.AuthInfos[0].Name != "prod-user" {
			t.Errorf("expected exactly 1 user 'prod-user', got user count: %d", len(gotConfig.AuthInfos))
		}
	})

	t.Run("minify is false", func(t *testing.T) {
		configContent := []byte("kubeconfig:\n  minify: false\n")
		if err := os.WriteFile(filepath.Join(homeDir, ".config", "ksw", "config.yaml"), configContent, 0600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		gotBytes, err := generateKubeconfig(kubeconfigPath, "prod-cluster")
		if err != nil {
			t.Fatalf("generateKubeconfig() error = %v", err)
		}

		var gotConfig apiv1.Config
		if err := yaml.Unmarshal(gotBytes, &gotConfig); err != nil {
			t.Fatalf("Failed to unmarshal generated kubeconfig: %v", err)
		}

		if gotConfig.CurrentContext != "prod-cluster" {
			t.Errorf("expected CurrentContext 'prod-cluster', got '%s'", gotConfig.CurrentContext)
		}

		if len(gotConfig.Contexts) != 2 {
			t.Errorf("expected 2 contexts, got %d", len(gotConfig.Contexts))
		}

		if len(gotConfig.Clusters) != 2 {
			t.Errorf("expected 2 clusters, got %d", len(gotConfig.Clusters))
		}

		if len(gotConfig.AuthInfos) != 2 {
			t.Errorf("expected 2 users, got %d", len(gotConfig.AuthInfos))
		}
	})
}
