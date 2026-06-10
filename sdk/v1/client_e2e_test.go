//go:build e2e
// +build e2e

package v1

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetCluster_RealOCM tests GetCluster against real OCM staging environment
func TestGetCluster_RealOCM(t *testing.T) {
	// Require credentials
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	// Use staging by default for safety
	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	// Create real client
	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	// Test with real OCM API
	cluster, err := client.GetCluster(testClusterID)

	// Assertions
	assert.NoError(t, err, "GetCluster should not return an error")
	assert.NotNil(t, cluster, "Cluster should not be nil")
	if cluster != nil {
		assert.Equal(t, testClusterID, cluster.ID(), "Cluster ID should match")
		assert.NotEmpty(t, cluster.Name(), "Cluster name should not be empty")
		assertClusterValid(t, cluster)
	}
}

// TestGetClusterAnyStatus_RealOCM tests GetClusterAnyStatus against real OCM
func TestGetClusterAnyStatus_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	cluster, err := client.GetClusterAnyStatus(testClusterID)

	assert.NoError(t, err, "GetClusterAnyStatus should not return an error")
	assert.NotNil(t, cluster, "Cluster should not be nil")
	if cluster != nil {
		assert.Equal(t, testClusterID, cluster.ID(), "Cluster ID should match")
		assertClusterValid(t, cluster)
	}
}

// TestGetClusters_RealOCM tests batch cluster retrieval
func TestGetClusters_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	// Test with single cluster
	clusters, err := client.GetClusters([]string{testClusterID})

	assert.NoError(t, err, "GetClusters should not return an error")
	assert.NotEmpty(t, clusters, "Clusters list should not be empty")
	if len(clusters) > 0 {
		assert.NotNil(t, clusters[0], "First cluster should not be nil")
		if clusters[0] != nil {
			assertClusterValid(t, clusters[0])
		}
	}
}

// TestIsClusterCCS_RealOCM tests CCS flag detection
func TestIsClusterCCS_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	cluster, err := client.GetCluster(testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	isCCS, err := client.IsClusterCCS(cluster)

	assert.NoError(t, err, "IsClusterCCS should not return an error")
	// Don't assert the actual value as it depends on the test cluster
	// Just verify the method executes without error
	t.Logf("Cluster %s CCS status: %v", testClusterID, isCCS)
}

// TestIsHostedCluster_RealOCM tests Hypershift/HCP detection
func TestIsHostedCluster_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	cluster, err := client.GetCluster(testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	isHosted, err := client.IsHostedCluster(cluster)

	assert.NoError(t, err, "IsHostedCluster should not return an error")
	t.Logf("Cluster %s hosted status: %v", testClusterID, isHosted)
}

// TestGetSubscription_RealOCM tests subscription lookup
func TestGetSubscription_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	// Get subscription using cluster ID
	subscription, err := client.GetSubscription(testClusterID)

	// Subscription lookup may fail if cluster doesn't have subscription
	// or if we don't have proper permissions
	if err != nil {
		t.Logf("GetSubscription failed (may be expected): %v", err)
	} else {
		assert.NotNil(t, subscription, "Subscription should not be nil")
		if subscription != nil {
			assertSubscriptionValid(t, subscription)
		}
	}
}

// TestGetOrgFromClusterID_RealOCM tests organization lookup from cluster
func TestGetOrgFromClusterID_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	orgID, err := client.GetOrgFromClusterID(testClusterID)

	// May fail depending on permissions
	if err != nil {
		t.Logf("GetOrgFromClusterID failed (may be expected): %v", err)
	} else {
		assert.NotEmpty(t, orgID, "Organization ID should not be empty")
		t.Logf("Cluster %s belongs to organization: %s", testClusterID, orgID)
	}
}

// TestGetAWSAccountIdForCluster_RealOCM tests AWS account ID extraction
func TestGetAWSAccountIdForCluster_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)
	defer client.Close()

	cluster, err := client.GetCluster(testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	accountID, err := client.GetAWSAccountIdForCluster(cluster.ID())

	// May fail if cluster is not AWS-based or lacks AWS configuration
	if err != nil {
		t.Logf("GetAWSAccountIdForCluster failed (may be expected for non-AWS cluster): %v", err)
	} else {
		assert.NotEmpty(t, accountID, "AWS account ID should not be empty")
		assert.Len(t, accountID, 12, "AWS account ID should be 12 digits")
		t.Logf("Cluster %s AWS account ID: %s", testClusterID, accountID)
	}
}

// TestConnectionLifecycle_RealOCM tests the automatic connection lifecycle
func TestConnectionLifecycle_RealOCM(t *testing.T) {
	if os.Getenv("OCM_TOKEN") == "" {
		t.Skip("Skipping E2E test: OCM_TOKEN not set")
	}

	ocmURL := getEnvOrDefault("OCM_URL", "staging")
	testClusterID := os.Getenv("TEST_CLUSTER_ID")
	if testClusterID == "" {
		t.Skip("Skipping: TEST_CLUSTER_ID not set")
	}

	t.Setenv("OCM_URL", ocmURL)
	client := setupTestClient(t, ocmURL)

	// Test that connection is auto-created on first method call
	cluster, err := client.GetCluster(testClusterID)
	require.NoError(t, err, "First method call should auto-create connection")
	assert.NotNil(t, cluster, "Cluster should not be nil")

	// Test that subsequent calls reuse the same connection
	cluster2, err := client.GetCluster(testClusterID)
	require.NoError(t, err, "Subsequent calls should reuse connection")
	assert.NotNil(t, cluster2, "Cluster should not be nil")

	// Test closing connection
	err = client.Close()
	assert.NoError(t, err, "Close should succeed")
}
