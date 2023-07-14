package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

func minifyConfig(c apiv1.Config, contextName string) (*apiv1.Config, error) {
	config := apiv1.Config{
		Kind:           c.Kind,
		APIVersion:     c.APIVersion,
		Preferences:    c.Preferences,
		Extensions:     c.Extensions,
		CurrentContext: contextName,
	}

	for _, context := range c.Contexts {
		if context.Name == contextName {
			config.Contexts = []apiv1.NamedContext{context}

			for _, authInfo := range c.AuthInfos {
				if authInfo.Name == context.Context.AuthInfo {
					config.AuthInfos = []apiv1.NamedAuthInfo{authInfo}
				}
			}

			for _, cluster := range c.Clusters {
				if cluster.Name == context.Context.Cluster {
					config.Clusters = []apiv1.NamedCluster{cluster}
				}
			}
		}
	}

	if len(config.Contexts) == 0 {
		return nil, fmt.Errorf("context not found")
	}

	return &config, nil
}

func getOriginalKubeconfigPath() string {
	var kubeconfigPath string // TODO add condition to use original kubeconfig from cli flags
	if path := os.Getenv("KSW_KUBECONFIG_ORIGINAL"); path != "" {
		kubeconfigPath = path
	} else if path := os.Getenv("KUBECONFIG"); path != "" {
		kubeconfigPath = path
	} else {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	return kubeconfigPath
}

func generateKubeconfig(path string, contextName string) ([]byte, error) {
	sourceBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config apiv1.Config

	if err := yaml.Unmarshal(sourceBytes, &config); err != nil {
		return nil, err
	}

	miniConfig, err := minifyConfig(config, contextName)
	if err != nil {
		return nil, err
	}

	bytes, err := yaml.Marshal(miniConfig)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
