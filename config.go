package main

import (
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

// KswConfig represents the application configuration.
type KswConfig struct {
	Kubeconfig KubeconfigConfig `json:"kubeconfig" yaml:"kubeconfig"`
}

// KubeconfigConfig holds configuration related to kubeconfig minification.
type KubeconfigConfig struct {
	Minify bool `json:"minify" yaml:"minify"`
}

var userHomeDir = os.UserHomeDir

// loadConfig loads the configuration from ~/.config/ksw/config.yaml or falling back to ~/.ksw.yaml.
func loadConfig() KswConfig {
	home, err := userHomeDir()
	if err != nil {
		return KswConfig{
			Kubeconfig: KubeconfigConfig{
				Minify: false,
			},
		}
	}

	return loadConfigFromHome(home)
}

// loadConfigFromHome loads the configuration relative to a specific home directory.
func loadConfigFromHome(home string) KswConfig {
	cfg := KswConfig{
		Kubeconfig: KubeconfigConfig{
			Minify: false,
		},
	}

	primaryPath := filepath.Join(home, ".config", "ksw", "config.yaml")
	fallbackPath := filepath.Join(home, ".ksw.yaml")

	var configBytes []byte

	var readErr error

	if _, err := os.Stat(primaryPath); err == nil {
		configBytes, readErr = os.ReadFile(primaryPath)
	} else if _, err := os.Stat(fallbackPath); err == nil {
		configBytes, readErr = os.ReadFile(fallbackPath)
	} else {
		return cfg
	}

	if readErr != nil {
		return cfg
	}

	var parsedCfg KswConfig

	parsedCfg.Kubeconfig.Minify = false

	if err := yaml.Unmarshal(configBytes, &parsedCfg); err != nil {
		return cfg
	}

	return parsedCfg
}
