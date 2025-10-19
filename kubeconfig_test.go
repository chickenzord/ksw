package main

import (
	"os"
	"path/filepath"
	"testing"

	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestMinifyConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      apiv1.Config
		contextName string
		wantErr     bool
		wantContext string
		wantCluster string
		wantUser    string
	}{
		{
			name: "valid context",
			config: apiv1.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Contexts: []apiv1.NamedContext{
					{
						Name: "prod-cluster",
						Context: apiv1.Context{
							Cluster:  "prod",
							AuthInfo: "prod-user",
						},
					},
					{
						Name: "dev-cluster",
						Context: apiv1.Context{
							Cluster:  "dev",
							AuthInfo: "dev-user",
						},
					},
				},
				Clusters: []apiv1.NamedCluster{
					{Name: "prod", Cluster: apiv1.Cluster{Server: "https://prod.example.com"}},
					{Name: "dev", Cluster: apiv1.Cluster{Server: "https://dev.example.com"}},
				},
				AuthInfos: []apiv1.NamedAuthInfo{
					{Name: "prod-user", AuthInfo: apiv1.AuthInfo{Token: "prod-token"}},
					{Name: "dev-user", AuthInfo: apiv1.AuthInfo{Token: "dev-token"}},
				},
			},
			contextName: "prod-cluster",
			wantErr:     false,
			wantContext: "prod-cluster",
			wantCluster: "prod",
			wantUser:    "prod-user",
		},
		{
			name: "context not found",
			config: apiv1.Config{
				Contexts: []apiv1.NamedContext{
					{Name: "prod-cluster", Context: apiv1.Context{Cluster: "prod", AuthInfo: "prod-user"}},
				},
				Clusters:  []apiv1.NamedCluster{{Name: "prod"}},
				AuthInfos: []apiv1.NamedAuthInfo{{Name: "prod-user"}},
			},
			contextName: "nonexistent",
			wantErr:     true,
		},
		{
			name: "minified config only contains specified context",
			config: apiv1.Config{
				Contexts: []apiv1.NamedContext{
					{Name: "ctx1", Context: apiv1.Context{Cluster: "c1", AuthInfo: "u1"}},
					{Name: "ctx2", Context: apiv1.Context{Cluster: "c2", AuthInfo: "u2"}},
				},
				Clusters:  []apiv1.NamedCluster{{Name: "c1"}, {Name: "c2"}},
				AuthInfos: []apiv1.NamedAuthInfo{{Name: "u1"}, {Name: "u2"}},
			},
			contextName: "ctx1",
			wantErr:     false,
			wantContext: "ctx1",
			wantCluster: "c1",
			wantUser:    "u1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := minifyConfig(tt.config, tt.contextName)
			if (err != nil) != tt.wantErr {
				t.Errorf("minifyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.CurrentContext != tt.wantContext {
				t.Errorf("minifyConfig() CurrentContext = %v, want %v", got.CurrentContext, tt.wantContext)
			}

			if len(got.Contexts) != 1 {
				t.Errorf("minifyConfig() has %d contexts, want 1", len(got.Contexts))
			}

			if len(got.Clusters) != 1 {
				t.Errorf("minifyConfig() has %d clusters, want 1", len(got.Clusters))
			}

			if len(got.AuthInfos) != 1 {
				t.Errorf("minifyConfig() has %d authInfos, want 1", len(got.AuthInfos))
			}

			if got.Contexts[0].Name != tt.wantContext {
				t.Errorf("minifyConfig() context name = %v, want %v", got.Contexts[0].Name, tt.wantContext)
			}

			if got.Clusters[0].Name != tt.wantCluster {
				t.Errorf("minifyConfig() cluster name = %v, want %v", got.Clusters[0].Name, tt.wantCluster)
			}

			if got.AuthInfos[0].Name != tt.wantUser {
				t.Errorf("minifyConfig() user name = %v, want %v", got.AuthInfos[0].Name, tt.wantUser)
			}
		})
	}
}

func TestListContexts(t *testing.T) {
	// Create a temporary kubeconfig file
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfigContent := `apiVersion: v1
kind: Config
contexts:
- name: prod-cluster
  context:
    cluster: prod
    user: prod-user
- name: dev-cluster
  context:
    cluster: dev
    user: dev-user
- name: staging-cluster
  context:
    cluster: staging
    user: staging-user
clusters:
- name: prod
  cluster:
    server: https://prod.example.com
- name: dev
  cluster:
    server: https://dev.example.com
- name: staging
  cluster:
    server: https://staging.example.com
users:
- name: prod-user
  user:
    token: prod-token
- name: dev-user
  user:
    token: dev-token
- name: staging-user
  user:
    token: staging-token
`

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0600); err != nil {
		t.Fatalf("Failed to create test kubeconfig: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    []string
		wantErr bool
	}{
		{
			name:    "valid kubeconfig",
			path:    kubeconfigPath,
			want:    []string{"prod-cluster", "dev-cluster", "staging-cluster"},
			wantErr: false,
		},
		{
			name:    "nonexistent file",
			path:    filepath.Join(tmpDir, "nonexistent"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listContexts(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("listContexts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("listContexts() returned %d contexts, want %d", len(got), len(tt.want))
				return
			}

			// Check that all expected contexts are present
			contextMap := make(map[string]bool)
			for _, ctx := range got {
				contextMap[ctx] = true
			}

			for _, wantCtx := range tt.want {
				if !contextMap[wantCtx] {
					t.Errorf("listContexts() missing context %q", wantCtx)
				}
			}
		})
	}
}

func TestGenerateKubeconfig(t *testing.T) {
	// Create a temporary kubeconfig file
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

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

	tests := []struct {
		name        string
		sourcePath  string
		contextName string
		wantErr     bool
	}{
		{
			name:        "valid context",
			sourcePath:  kubeconfigPath,
			contextName: "prod-cluster",
			wantErr:     false,
		},
		{
			name:        "nonexistent context",
			sourcePath:  kubeconfigPath,
			contextName: "nonexistent",
			wantErr:     true,
		},
		{
			name:        "nonexistent file",
			sourcePath:  filepath.Join(tmpDir, "nonexistent"),
			contextName: "prod-cluster",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateKubeconfig(tt.sourcePath, tt.contextName)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateKubeconfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got) == 0 {
				t.Error("generateKubeconfig() returned empty byte slice")
			}
		})
	}
}

func TestGetOriginalKubeconfigPath(t *testing.T) {
	// Save original env vars
	origKswOriginal := os.Getenv("KSW_KUBECONFIG_ORIGINAL")
	origKubeconfig := os.Getenv("KUBECONFIG")
	origHome := os.Getenv("HOME")

	// Restore after test
	defer func() {
		os.Setenv("KSW_KUBECONFIG_ORIGINAL", origKswOriginal)
		os.Setenv("KUBECONFIG", origKubeconfig)
		os.Setenv("HOME", origHome)
	}()

	tests := []struct {
		name              string
		kswOriginal       string
		kubeconfig        string
		home              string
		expectedContains  string
		setKswOriginal    bool
		setKubeconfig     bool
	}{
		{
			name:             "KSW_KUBECONFIG_ORIGINAL set",
			kswOriginal:      "/custom/ksw/config",
			setKswOriginal:   true,
			expectedContains: "/custom/ksw/config",
		},
		{
			name:             "KUBECONFIG set, KSW_KUBECONFIG_ORIGINAL not set",
			kubeconfig:       "/custom/kube/config",
			setKubeconfig:    true,
			expectedContains: "/custom/kube/config",
		},
		{
			name:             "default ~/.kube/config",
			home:             "/home/testuser",
			expectedContains: "/home/testuser/.kube/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			os.Unsetenv("KSW_KUBECONFIG_ORIGINAL")
			os.Unsetenv("KUBECONFIG")

			// Set env vars as needed
			if tt.setKswOriginal {
				os.Setenv("KSW_KUBECONFIG_ORIGINAL", tt.kswOriginal)
			}
			if tt.setKubeconfig {
				os.Setenv("KUBECONFIG", tt.kubeconfig)
			}
			if tt.home != "" {
				os.Setenv("HOME", tt.home)
			}

			got := getOriginalKubeconfigPath()
			if got != tt.expectedContains {
				t.Errorf("getOriginalKubeconfigPath() = %v, want %v", got, tt.expectedContains)
			}
		})
	}
}
