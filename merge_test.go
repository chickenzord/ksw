package main

import (
	"reflect"
	"testing"

	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestComputeKubeconfigDiff(t *testing.T) {
	orig := apiv1.Config{
		Contexts: []apiv1.NamedContext{
			{Name: "c1", Context: apiv1.Context{Cluster: "cluster1", AuthInfo: "user1"}},
			{Name: "c2", Context: apiv1.Context{Cluster: "cluster2", AuthInfo: "user2"}},
		},
		Clusters: []apiv1.NamedCluster{
			{Name: "cluster1", Cluster: apiv1.Cluster{Server: "https://c1.example.com"}},
			{Name: "cluster2", Cluster: apiv1.Cluster{Server: "https://c2.example.com"}},
		},
		AuthInfos: []apiv1.NamedAuthInfo{
			{Name: "user1", AuthInfo: apiv1.AuthInfo{Token: "token1"}},
			{Name: "user2", AuthInfo: apiv1.AuthInfo{Token: "token2"}},
		},
	}

	t.Run("no changes", func(t *testing.T) {
		temp := orig
		diff := computeKubeconfigDiff(orig, temp, false)

		if diff.HasChanges() {
			t.Errorf("expected no changes, got changes")
		}
	})

	t.Run("additions", func(t *testing.T) {
		temp := orig
		temp.Contexts = append(temp.Contexts, apiv1.NamedContext{Name: "c3", Context: apiv1.Context{Cluster: "cluster3", AuthInfo: "user3"}})
		temp.Clusters = append(temp.Clusters, apiv1.NamedCluster{Name: "cluster3", Cluster: apiv1.Cluster{Server: "https://c3.example.com"}})
		temp.AuthInfos = append(temp.AuthInfos, apiv1.NamedAuthInfo{Name: "user3", AuthInfo: apiv1.AuthInfo{Token: "token3"}})

		diff := computeKubeconfigDiff(orig, temp, false)

		if !diff.HasChanges() {
			t.Fatalf("expected changes, got none")
		}

		if len(diff.ContextsAdded) != 1 || diff.ContextsAdded[0].Name != "c3" {
			t.Errorf("expected 1 context added (c3), got %v", diff.ContextsAdded)
		}

		if len(diff.ClustersAdded) != 1 || diff.ClustersAdded[0].Name != "cluster3" {
			t.Errorf("expected 1 cluster added (cluster3), got %v", diff.ClustersAdded)
		}

		if len(diff.UsersAdded) != 1 || diff.UsersAdded[0].Name != "user3" {
			t.Errorf("expected 1 user added (user3), got %v", diff.UsersAdded)
		}
	})

	t.Run("modifications", func(t *testing.T) {
		temp := orig
		// Modify context c2 to point to user3
		temp.Contexts = []apiv1.NamedContext{
			{Name: "c1", Context: apiv1.Context{Cluster: "cluster1", AuthInfo: "user1"}},
			{Name: "c2", Context: apiv1.Context{Cluster: "cluster2", AuthInfo: "user3"}},
		}
		// Modify cluster1 server URL
		temp.Clusters = []apiv1.NamedCluster{
			{Name: "cluster1", Cluster: apiv1.Cluster{Server: "https://c1-new.example.com"}},
			{Name: "cluster2", Cluster: apiv1.Cluster{Server: "https://c2.example.com"}},
		}
		// Modify user2 token
		temp.AuthInfos = []apiv1.NamedAuthInfo{
			{Name: "user1", AuthInfo: apiv1.AuthInfo{Token: "token1"}},
			{Name: "user2", AuthInfo: apiv1.AuthInfo{Token: "token2-new"}},
		}

		diff := computeKubeconfigDiff(orig, temp, false)

		if !diff.HasChanges() {
			t.Fatalf("expected changes, got none")
		}

		if len(diff.ContextsModified) != 1 || diff.ContextsModified[0].Name != "c2" {
			t.Errorf("expected context c2 modified, got %v", diff.ContextsModified)
		}

		if len(diff.ClustersModified) != 1 || diff.ClustersModified[0].Name != "cluster1" {
			t.Errorf("expected cluster cluster1 modified, got %v", diff.ClustersModified)
		}

		if len(diff.UsersModified) != 1 || diff.UsersModified[0].Name != "user2" {
			t.Errorf("expected user user2 modified, got %v", diff.UsersModified)
		}
	})

	t.Run("deletions when not minified", func(t *testing.T) {
		temp := orig
		// Remove c2, cluster2, and user2
		temp.Contexts = []apiv1.NamedContext{orig.Contexts[0]}
		temp.Clusters = []apiv1.NamedCluster{orig.Clusters[0]}
		temp.AuthInfos = []apiv1.NamedAuthInfo{orig.AuthInfos[0]}

		diff := computeKubeconfigDiff(orig, temp, false)

		if !diff.HasChanges() {
			t.Fatalf("expected changes, got none")
		}

		if len(diff.ContextsDeleted) != 1 || diff.ContextsDeleted[0] != "c2" {
			t.Errorf("expected context c2 deleted, got %v", diff.ContextsDeleted)
		}

		if len(diff.ClustersDeleted) != 1 || diff.ClustersDeleted[0] != "cluster2" {
			t.Errorf("expected cluster cluster2 deleted, got %v", diff.ClustersDeleted)
		}

		if len(diff.UsersDeleted) != 1 || diff.UsersDeleted[0] != "user2" {
			t.Errorf("expected user user2 deleted, got %v", diff.UsersDeleted)
		}
	})

	t.Run("deletions ignored when minified", func(t *testing.T) {
		temp := orig
		// Remove c2, cluster2, and user2 (mimicking minified state)
		temp.Contexts = []apiv1.NamedContext{orig.Contexts[0]}
		temp.Clusters = []apiv1.NamedCluster{orig.Clusters[0]}
		temp.AuthInfos = []apiv1.NamedAuthInfo{orig.AuthInfos[0]}

		diff := computeKubeconfigDiff(orig, temp, true) // minified = true

		if diff.HasChanges() {
			t.Errorf("expected no changes (deletions ignored during minification), got %v", diff)
		}
	})
}

func TestApplyDiff(t *testing.T) {
	orig := apiv1.Config{
		Contexts: []apiv1.NamedContext{
			{Name: "c1", Context: apiv1.Context{Cluster: "cluster1", AuthInfo: "user1"}},
			{Name: "c3", Context: apiv1.Context{Cluster: "cluster3", AuthInfo: "user3"}},
		},
		Clusters: []apiv1.NamedCluster{
			{Name: "cluster1", Cluster: apiv1.Cluster{Server: "https://c1.example.com"}},
			{Name: "cluster3", Cluster: apiv1.Cluster{Server: "https://c3.example.com"}},
		},
		AuthInfos: []apiv1.NamedAuthInfo{
			{Name: "user1", AuthInfo: apiv1.AuthInfo{Token: "token1"}},
			{Name: "user3", AuthInfo: apiv1.AuthInfo{Token: "token3"}},
		},
	}

	diff := KubeconfigDiff{
		ContextsAdded: []apiv1.NamedContext{
			{Name: "c2", Context: apiv1.Context{Cluster: "cluster2", AuthInfo: "user2"}},
		},
		ContextsModified: []apiv1.NamedContext{
			{Name: "c1", Context: apiv1.Context{Cluster: "cluster1-new", AuthInfo: "user1-new"}},
		},
		ContextsDeleted: []string{"c3"},

		ClustersAdded: []apiv1.NamedCluster{
			{Name: "cluster2", Cluster: apiv1.Cluster{Server: "https://c2.example.com"}},
		},
		ClustersModified: []apiv1.NamedCluster{
			{Name: "cluster1", Cluster: apiv1.Cluster{Server: "https://c1-new.example.com"}},
		},
		ClustersDeleted: []string{"cluster3"},

		UsersAdded: []apiv1.NamedAuthInfo{
			{Name: "user2", AuthInfo: apiv1.AuthInfo{Token: "token2"}},
		},
		UsersModified: []apiv1.NamedAuthInfo{
			{Name: "user1", AuthInfo: apiv1.AuthInfo{Token: "token1-new"}},
		},
		UsersDeleted: []string{"user3"},
	}

	got := applyDiff(orig, diff)

	// Check Contexts
	wantContexts := []apiv1.NamedContext{
		{Name: "c1", Context: apiv1.Context{Cluster: "cluster1-new", AuthInfo: "user1-new"}},
		{Name: "c2", Context: apiv1.Context{Cluster: "cluster2", AuthInfo: "user2"}},
	}

	if !reflect.DeepEqual(got.Contexts, wantContexts) {
		t.Errorf("applyDiff Contexts = %v, want %v", got.Contexts, wantContexts)
	}

	// Check Clusters
	wantClusters := []apiv1.NamedCluster{
		{Name: "cluster1", Cluster: apiv1.Cluster{Server: "https://c1-new.example.com"}},
		{Name: "cluster2", Cluster: apiv1.Cluster{Server: "https://c2.example.com"}},
	}

	if !reflect.DeepEqual(got.Clusters, wantClusters) {
		t.Errorf("applyDiff Clusters = %v, want %v", got.Clusters, wantClusters)
	}

	// Check Users
	wantUsers := []apiv1.NamedAuthInfo{
		{Name: "user1", AuthInfo: apiv1.AuthInfo{Token: "token1-new"}},
		{Name: "user2", AuthInfo: apiv1.AuthInfo{Token: "token2"}},
	}

	if !reflect.DeepEqual(got.AuthInfos, wantUsers) {
		t.Errorf("applyDiff Users = %v, want %v", got.AuthInfos, wantUsers)
	}
}
