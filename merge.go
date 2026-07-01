package main

import (
	"fmt"
	"os"
	"reflect"
	"slices"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// KubeconfigDiff represents the added, modified, and deleted entities between configs.
type KubeconfigDiff struct {
	ContextsAdded    []apiv1.NamedContext
	ContextsModified []apiv1.NamedContext
	ContextsDeleted  []string

	ClustersAdded    []apiv1.NamedCluster
	ClustersModified []apiv1.NamedCluster
	ClustersDeleted  []string

	UsersAdded    []apiv1.NamedAuthInfo
	UsersModified []apiv1.NamedAuthInfo
	UsersDeleted  []string
}

// HasChanges returns true if there are any differences.
func (d KubeconfigDiff) HasChanges() bool {
	return len(d.ContextsAdded) > 0 || len(d.ContextsModified) > 0 || len(d.ContextsDeleted) > 0 ||
		len(d.ClustersAdded) > 0 || len(d.ClustersModified) > 0 || len(d.ClustersDeleted) > 0 ||
		len(d.UsersAdded) > 0 || len(d.UsersModified) > 0 || len(d.UsersDeleted) > 0
}

func contextsMap(contexts []apiv1.NamedContext) map[string]apiv1.Context {
	m := make(map[string]apiv1.Context)

	for _, c := range contexts {
		m[c.Name] = c.Context
	}

	return m
}

func clustersMap(clusters []apiv1.NamedCluster) map[string]apiv1.Cluster {
	m := make(map[string]apiv1.Cluster)

	for _, c := range clusters {
		m[c.Name] = c.Cluster
	}

	return m
}

func usersMap(authInfos []apiv1.NamedAuthInfo) map[string]apiv1.AuthInfo {
	m := make(map[string]apiv1.AuthInfo)

	for _, u := range authInfos {
		m[u.Name] = u.AuthInfo
	}

	return m
}

// computeKubeconfigDiff calculates the differences between original and temporary configs.
func computeKubeconfigDiff(orig, temp apiv1.Config, minified bool) KubeconfigDiff {
	var diff KubeconfigDiff

	origCtxs := contextsMap(orig.Contexts)
	tempCtxs := contextsMap(temp.Contexts)

	origClusters := clustersMap(orig.Clusters)
	tempClusters := clustersMap(temp.Clusters)

	origUsers := usersMap(orig.AuthInfos)
	tempUsers := usersMap(temp.AuthInfos)

	// Compare contexts
	for name, tempCtx := range tempCtxs {
		if origCtx, exists := origCtxs[name]; !exists {
			diff.ContextsAdded = append(diff.ContextsAdded, apiv1.NamedContext{Name: name, Context: tempCtx})
		} else if !reflect.DeepEqual(origCtx, tempCtx) {
			diff.ContextsModified = append(diff.ContextsModified, apiv1.NamedContext{Name: name, Context: tempCtx})
		}
	}

	if !minified {
		for name := range origCtxs {
			if _, exists := tempCtxs[name]; !exists {
				diff.ContextsDeleted = append(diff.ContextsDeleted, name)
			}
		}
	}

	// Compare clusters
	for name, tempCluster := range tempClusters {
		if origCluster, exists := origClusters[name]; !exists {
			diff.ClustersAdded = append(diff.ClustersAdded, apiv1.NamedCluster{Name: name, Cluster: tempCluster})
		} else if !reflect.DeepEqual(origCluster, tempCluster) {
			diff.ClustersModified = append(diff.ClustersModified, apiv1.NamedCluster{Name: name, Cluster: tempCluster})
		}
	}

	if !minified {
		for name := range origClusters {
			if _, exists := tempClusters[name]; !exists {
				diff.ClustersDeleted = append(diff.ClustersDeleted, name)
			}
		}
	}

	// Compare users
	for name, tempUser := range tempUsers {
		if origUser, exists := origUsers[name]; !exists {
			diff.UsersAdded = append(diff.UsersAdded, apiv1.NamedAuthInfo{Name: name, AuthInfo: tempUser})
		} else if !reflect.DeepEqual(origUser, tempUser) {
			diff.UsersModified = append(diff.UsersModified, apiv1.NamedAuthInfo{Name: name, AuthInfo: tempUser})
		}
	}

	if !minified {
		for name := range origUsers {
			if _, exists := tempUsers[name]; !exists {
				diff.UsersDeleted = append(diff.UsersDeleted, name)
			}
		}
	}

	return diff
}

// applyDiff merges the selected differences back into a config.
func applyDiff(orig apiv1.Config, diff KubeconfigDiff) apiv1.Config {
	ctxMap := contextsMap(orig.Contexts)
	clusterMap := clustersMap(orig.Clusters)
	userMap := usersMap(orig.AuthInfos)

	// Apply contexts
	for _, x := range diff.ContextsAdded {
		ctxMap[x.Name] = x.Context
	}

	for _, x := range diff.ContextsModified {
		ctxMap[x.Name] = x.Context
	}

	for _, name := range diff.ContextsDeleted {
		delete(ctxMap, name)
	}

	// Apply clusters
	for _, x := range diff.ClustersAdded {
		clusterMap[x.Name] = x.Cluster
	}

	for _, x := range diff.ClustersModified {
		clusterMap[x.Name] = x.Cluster
	}

	for _, name := range diff.ClustersDeleted {
		delete(clusterMap, name)
	}

	// Apply users
	for _, x := range diff.UsersAdded {
		userMap[x.Name] = x.AuthInfo
	}

	for _, x := range diff.UsersModified {
		userMap[x.Name] = x.AuthInfo
	}

	for _, name := range diff.UsersDeleted {
		delete(userMap, name)
	}

	// Rebuild slices sorted alphabetically
	var (
		contexts     []apiv1.NamedContext
		contextsKeys []string
	)

	for k := range ctxMap {
		contextsKeys = append(contextsKeys, k)
	}

	slices.Sort(contextsKeys)

	for _, k := range contextsKeys {
		contexts = append(contexts, apiv1.NamedContext{Name: k, Context: ctxMap[k]})
	}

	var (
		clusters     []apiv1.NamedCluster
		clustersKeys []string
	)

	for k := range clusterMap {
		clustersKeys = append(clustersKeys, k)
	}

	slices.Sort(clustersKeys)

	for _, k := range clustersKeys {
		clusters = append(clusters, apiv1.NamedCluster{Name: k, Cluster: clusterMap[k]})
	}

	var (
		users     []apiv1.NamedAuthInfo
		usersKeys []string
	)

	for k := range userMap {
		usersKeys = append(usersKeys, k)
	}

	slices.Sort(usersKeys)

	for _, k := range usersKeys {
		users = append(users, apiv1.NamedAuthInfo{Name: k, AuthInfo: userMap[k]})
	}

	orig.Contexts = contexts
	orig.Clusters = clusters
	orig.AuthInfos = users

	return orig
}

// mergeOnExit loads both configs, identifies changes, shows an interactive prompt, and applies them.
func mergeOnExit(originalPath, tempPath string, minified bool) error {
	origBytes, err := os.ReadFile(originalPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read original kubeconfig: %w", err)
	}

	var origConfig apiv1.Config

	if len(origBytes) > 0 {
		if err := yaml.Unmarshal(origBytes, &origConfig); err != nil {
			return fmt.Errorf("failed to unmarshal original kubeconfig: %w", err)
		}
	}

	tempBytes, err := os.ReadFile(tempPath)
	if err != nil {
		return fmt.Errorf("failed to read temporary kubeconfig: %w", err)
	}

	var tempConfig apiv1.Config

	if err := yaml.Unmarshal(tempBytes, &tempConfig); err != nil {
		return fmt.Errorf("failed to unmarshal temporary kubeconfig: %w", err)
	}

	diff := computeKubeconfigDiff(origConfig, tempConfig, minified)
	if !diff.HasChanges() {
		return nil
	}

	selectedDiff, err := selectChanges(diff, originalPath, tempPath)
	if err != nil {
		return fmt.Errorf("error selecting changes: %w", err)
	}

	if !selectedDiff.HasChanges() {
		fmt.Println("No changes selected to merge.")

		return nil
	}

	latestOrigBytes, err := os.ReadFile(originalPath)

	var latestOrigConfig apiv1.Config

	if err == nil {
		if err := yaml.Unmarshal(latestOrigBytes, &latestOrigConfig); err != nil {
			return fmt.Errorf("failed to unmarshal latest original kubeconfig: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read latest original kubeconfig: %w", err)
	}

	mergedConfig := applyDiff(latestOrigConfig, selectedDiff)

	mergedBytes, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal merged kubeconfig: %w", err)
	}

	if err := os.WriteFile(originalPath, mergedBytes, 0600); err != nil {
		return fmt.Errorf("failed to write original kubeconfig: %w", err)
	}

	fmt.Println("Selected changes successfully applied back to original kubeconfig.")

	return nil
}
