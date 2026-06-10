//go:build e2e
// +build e2e

package v1

import (
	"os"
	"testing"

	"github.com/petrkotas/srelib/internal/i1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2E Test Mode Detection
// These tests can run in two modes:
// 1. Mock Mode (local development): Uses mock OCM server with fixtures
// 2. Real Mode (CI): Uses real OCM staging/production with actual credentials
//
// The mode is automatically detected based on OCM_TOKEN presence:
// - If OCM_TOKEN is set → Real OCM mode
// - If OCM_TOKEN is empty → Mock server mode
//
// This allows:
// - Developers to run E2E tests locally without credentials
// - CI to run the same tests against real OCM for validation

// e2eTestContext holds the test setup (either mock or real)
type e2eTestContext struct {
	client        *i1.Client
	mockServer    *MockOCMServer
	testClusterID string
	isRealOCM     bool
}

// setupE2ETest creates test context (mock or real OCM based on credentials)
func setupE2ETest(t *testing.T) *e2eTestContext {
	ctx := &e2eTestContext{}

	// Check if we have real OCM credentials
	ocmToken := os.Getenv("OCM_TOKEN")
	ctx.isRealOCM = (ocmToken != "")

	if ctx.isRealOCM {
		// Real OCM mode - use credentials
		t.Log("Running E2E test in REAL OCM mode")

		ocmURL := getEnvOrDefault("OCM_URL", "staging")
		ctx.testClusterID = os.Getenv("TEST_CLUSTER_ID")

		if ctx.testClusterID == "" {
			t.Skip("Skipping real OCM test: TEST_CLUSTER_ID not set")
		}

		t.Setenv("OCM_URL", ocmURL)
		ctx.client = setupTestClient(t)
	} else {
		// Mock mode - use fixtures
		t.Log("Running E2E test in MOCK mode (no OCM_TOKEN found)")

		ctx.mockServer = NewMockOCMServer(t)

		// Load test cluster fixture
		clusterID, err := ctx.mockServer.LoadClusterFromFixture("cluster_osd.json")
		require.NoError(t, err, "Failed to load cluster fixture")
		ctx.testClusterID = clusterID

		// Configure client to use mock server
		// Use a valid JWT structure (header.payload.signature in base64)
		t.Setenv("OCM_URL", ctx.mockServer.Server.URL)
		t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
		ctx.client = setupTestClient(t)
	}

	return ctx
}

// Close cleans up test resources
func (ctx *e2eTestContext) Close() {
	if ctx.client != nil {
		ctx.client.Close()
	}
	if ctx.mockServer != nil {
		ctx.mockServer.Close()
	}
}

// TestGetCluster_E2E tests GetCluster (works with both mock and real OCM)
func TestGetCluster_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	// Test GetCluster
	cluster, err := ctx.client.GetCluster(ctx.testClusterID)

	// Assertions
	assert.NoError(t, err, "GetCluster should not return an error")
	assert.NotNil(t, cluster, "Cluster should not be nil")
	if cluster != nil {
		assert.Equal(t, ctx.testClusterID, cluster.ID(), "Cluster ID should match")
		assert.NotEmpty(t, cluster.Name(), "Cluster name should not be empty")
		assertClusterValid(t, cluster)

		if ctx.isRealOCM {
			t.Logf("Real OCM - Cluster: %s (%s)", cluster.Name(), cluster.ID())
		} else {
			t.Logf("Mock - Cluster: %s (%s)", cluster.Name(), cluster.ID())
		}
	}
}

// TestGetClusterAnyStatus_E2E tests GetClusterAnyStatus (works with both mock and real OCM)
func TestGetClusterAnyStatus_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	cluster, err := ctx.client.GetClusterAnyStatus(ctx.testClusterID)

	assert.NoError(t, err, "GetClusterAnyStatus should not return an error")
	assert.NotNil(t, cluster, "Cluster should not be nil")
	if cluster != nil {
		assert.Equal(t, ctx.testClusterID, cluster.ID(), "Cluster ID should match")
		assertClusterValid(t, cluster)

		if ctx.isRealOCM {
			t.Logf("Real OCM - Cluster status: %s", cluster.State())
		}
	}
}

// TestGetClusters_E2E tests batch cluster retrieval (works with both mock and real OCM)
func TestGetClusters_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	// Test with single cluster
	clusters, err := ctx.client.GetClusters([]string{ctx.testClusterID})

	assert.NoError(t, err, "GetClusters should not return an error")
	assert.NotEmpty(t, clusters, "Clusters list should not be empty")
	if len(clusters) > 0 {
		assert.NotNil(t, clusters[0], "First cluster should not be nil")
		if clusters[0] != nil {
			assertClusterValid(t, clusters[0])
			t.Logf("Retrieved %d cluster(s) in batch", len(clusters))
		}
	}
}

// TestGetSubscription_E2E tests subscription lookup (works with both mock and real OCM)
func TestGetSubscription_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	// For mock mode, load subscription fixture
	if !ctx.isRealOCM {
		_, err := ctx.mockServer.LoadSubscriptionFromFixture("subscription_response.json")
		require.NoError(t, err, "Failed to load subscription fixture")
	}

	// Get subscription using cluster ID
	subscription, err := ctx.client.GetSubscription(ctx.testClusterID)

	if ctx.isRealOCM {
		// Real OCM: subscription lookup may fail depending on permissions
		if err != nil {
			t.Logf("GetSubscription failed (may be expected): %v", err)
		} else {
			assert.NotNil(t, subscription, "Subscription should not be nil")
			if subscription != nil {
				assertSubscriptionValid(t, subscription)
				t.Logf("Real OCM - Subscription ID: %s", subscription.ID())
			}
		}
	} else {
		// Mock mode: should succeed with fixture
		assert.NoError(t, err, "GetSubscription should not return an error in mock mode")
		assert.NotNil(t, subscription, "Subscription should not be nil")
		if subscription != nil {
			assertSubscriptionValid(t, subscription)
		}
	}
}

// TestGetAWSAccountIdForCluster_E2E tests AWS account ID extraction (works with both mock and real OCM)
func TestGetAWSAccountIdForCluster_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	cluster, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	accountID, err := ctx.client.GetAWSAccountIdForCluster(cluster.ID())

	// Both mock and real: may fail if cluster is not AWS-based
	if err != nil {
		t.Logf("GetAWSAccountIdForCluster failed (may be expected for non-AWS cluster): %v", err)
	} else {
		assert.NotEmpty(t, accountID, "AWS account ID should not be empty")
		assert.Len(t, accountID, 12, "AWS account ID should be 12 digits")

		if ctx.isRealOCM {
			t.Logf("Real OCM - Cluster %s AWS account ID: %s", ctx.testClusterID, accountID)
		} else {
			t.Logf("Mock - AWS account ID: %s", accountID)
		}
	}
}

// TestConnectionLifecycle_E2E tests the automatic connection lifecycle (works with both mock and real OCM)
func TestConnectionLifecycle_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	// Don't defer Close() here - we test it explicitly

	// Test that connection works on first method call
	cluster, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err, "First method call should work")
	assert.NotNil(t, cluster, "Cluster should not be nil")

	// Test that subsequent calls reuse the same connection
	cluster2, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err, "Subsequent calls should reuse connection")
	assert.NotNil(t, cluster2, "Cluster should not be nil")

	// Test closing connection
	err = ctx.client.Close()
	assert.NoError(t, err, "Close should succeed")

	// Clean up mock server if present
	if ctx.mockServer != nil {
		ctx.mockServer.Close()
	}

	if ctx.isRealOCM {
		t.Log("Real OCM - Connection lifecycle validated")
	} else {
		t.Log("Mock - Connection lifecycle validated")
	}
}

// TestGetCluster_CCS_E2E tests CCS cluster detection using fixture (mock mode only)
func TestGetCluster_CCS_E2E(t *testing.T) {
	if os.Getenv("OCM_TOKEN") != "" {
		t.Skip("Skipping CCS fixture test when running against real OCM")
	}

	ctx := &e2eTestContext{}
	ctx.mockServer = NewMockOCMServer(t)
	defer ctx.mockServer.Close()

	clusterID, err := ctx.mockServer.LoadClusterFromFixture("cluster_ccs.json")
	require.NoError(t, err, "Failed to load CCS cluster fixture")

	t.Setenv("OCM_URL", ctx.mockServer.Server.URL)
	t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
	ctx.client = setupTestClient(t)
	defer ctx.client.Close()

	cluster, err := ctx.client.GetCluster(clusterID)
	require.NoError(t, err, "GetCluster should not return an error")
	require.NotNil(t, cluster, "Cluster should not be nil")

	// Verify CCS flag from cluster object
	ccs, ok := cluster.GetCCS()
	assert.True(t, ok, "CCS should be present")
	assert.True(t, ccs.Enabled(), "CCS should be enabled")

	t.Logf("Successfully verified CCS cluster: %s (CCS enabled: %v)", cluster.Name(), ccs.Enabled())
}

// TestGetCluster_Hypershift_E2E tests Hypershift cluster detection using fixture (mock mode only)
func TestGetCluster_Hypershift_E2E(t *testing.T) {
	if os.Getenv("OCM_TOKEN") != "" {
		t.Skip("Skipping Hypershift fixture test when running against real OCM")
	}

	ctx := &e2eTestContext{}
	ctx.mockServer = NewMockOCMServer(t)
	defer ctx.mockServer.Close()

	clusterID, err := ctx.mockServer.LoadClusterFromFixture("cluster_hypershift.json")
	require.NoError(t, err, "Failed to load Hypershift cluster fixture")

	t.Setenv("OCM_URL", ctx.mockServer.Server.URL)
	t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
	ctx.client = setupTestClient(t)
	defer ctx.client.Close()

	cluster, err := ctx.client.GetCluster(clusterID)
	require.NoError(t, err, "GetCluster should not return an error")
	require.NotNil(t, cluster, "Cluster should not be nil")

	// Verify Hypershift flag from cluster object
	hypershift, ok := cluster.GetHypershift()
	assert.True(t, ok, "Hypershift should be present")
	assert.True(t, hypershift.Enabled(), "Hypershift should be enabled")

	t.Logf("Successfully verified Hypershift cluster: %s (Hypershift enabled: %v)", cluster.Name(), hypershift.Enabled())
}
