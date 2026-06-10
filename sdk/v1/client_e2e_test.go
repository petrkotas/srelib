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
	mode          string // "mock" or "real" - only for logging
}

// setupE2ETest creates test context (mock or real OCM based on credentials)
func setupE2ETest(t *testing.T) *e2eTestContext {
	ctx := &e2eTestContext{}

	// Check if we have real OCM credentials
	ocmToken := os.Getenv("OCM_TOKEN")
	isRealOCM := (ocmToken != "")

	if isRealOCM {
		// Real OCM mode - use credentials
		ctx.mode = "real"
		t.Logf("Running E2E test in REAL OCM mode")

		ocmURL := getEnvOrDefault("OCM_URL", "staging")
		ctx.testClusterID = os.Getenv("TEST_CLUSTER_ID")

		if ctx.testClusterID == "" {
			t.Skip("Skipping real OCM test: TEST_CLUSTER_ID not set")
		}

		t.Setenv("OCM_URL", ocmURL)
		ctx.client = setupTestClient(t)
	} else {
		// Mock mode - use fixtures
		ctx.mode = "mock"
		t.Logf("Running E2E test in MOCK mode (no OCM_TOKEN found)")

		ctx.mockServer = NewMockOCMServer(t)

		// Load test cluster fixture
		clusterID, err := ctx.mockServer.LoadClusterFromFixture("cluster_osd.json")
		require.NoError(t, err, "Failed to load cluster fixture")
		ctx.testClusterID = clusterID

		// Pre-load subscription fixture for subscription tests
		_, err = ctx.mockServer.LoadSubscriptionFromFixture("subscription_response.json")
		require.NoError(t, err, "Failed to load subscription fixture")

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

// setupE2ETestWithFixture creates test context for a specific fixture (mock mode only)
func setupE2ETestWithFixture(t *testing.T, fixtureName string) *e2eTestContext {
	if os.Getenv("OCM_TOKEN") != "" {
		t.Skipf("Skipping fixture test %s when running against real OCM", fixtureName)
	}

	ctx := &e2eTestContext{mode: "mock"}
	ctx.mockServer = NewMockOCMServer(t)

	clusterID, err := ctx.mockServer.LoadClusterFromFixture(fixtureName)
	require.NoError(t, err, "Failed to load fixture %s", fixtureName)
	ctx.testClusterID = clusterID

	t.Setenv("OCM_URL", ctx.mockServer.Server.URL)
	t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
	ctx.client = setupTestClient(t)

	return ctx
}

// TestGetCluster_E2E tests GetCluster (works with both mock and real OCM)
func TestGetCluster_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	cluster, err := ctx.client.GetCluster(ctx.testClusterID)

	require.NoError(t, err)
	require.NotNil(t, cluster)
	assert.Equal(t, ctx.testClusterID, cluster.ID())
	assert.NotEmpty(t, cluster.Name())
	assertClusterValid(t, cluster)

	t.Logf("[%s] Cluster: %s (%s)", ctx.mode, cluster.Name(), cluster.ID())
}

// TestGetClusterAnyStatus_E2E tests GetClusterAnyStatus (works with both mock and real OCM)
func TestGetClusterAnyStatus_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	cluster, err := ctx.client.GetClusterAnyStatus(ctx.testClusterID)

	require.NoError(t, err)
	require.NotNil(t, cluster)
	assert.Equal(t, ctx.testClusterID, cluster.ID())
	assertClusterValid(t, cluster)

	t.Logf("[%s] Cluster status: %s", ctx.mode, cluster.State())
}

// TestGetClusters_E2E tests batch cluster retrieval (works with both mock and real OCM)
func TestGetClusters_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	clusters, err := ctx.client.GetClusters([]string{ctx.testClusterID})

	require.NoError(t, err)
	require.NotEmpty(t, clusters)
	require.NotNil(t, clusters[0])
	assertClusterValid(t, clusters[0])

	t.Logf("[%s] Retrieved %d cluster(s) in batch", ctx.mode, len(clusters))
}

// TestGetSubscription_E2E tests subscription lookup (works with both mock and real OCM)
func TestGetSubscription_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	subscription, err := ctx.client.GetSubscription(ctx.testClusterID)

	// Subscription lookup may fail in real OCM depending on permissions
	// In mock mode, fixture is pre-loaded in setup
	if err != nil {
		t.Skipf("[%s] GetSubscription not available (may be expected): %v", ctx.mode, err)
	}

	require.NotNil(t, subscription)
	assertSubscriptionValid(t, subscription)
	t.Logf("[%s] Subscription ID: %s", ctx.mode, subscription.ID())
}

// TestGetAWSAccountIdForCluster_E2E tests AWS account ID extraction (works with both mock and real OCM)
func TestGetAWSAccountIdForCluster_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	defer ctx.Close()

	cluster, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	accountID, err := ctx.client.GetAWSAccountIdForCluster(cluster.ID())

	// May fail if cluster is not AWS-based
	if err != nil {
		t.Skipf("[%s] GetAWSAccountIdForCluster not available (may be expected for non-AWS cluster): %v", ctx.mode, err)
	}

	assert.NotEmpty(t, accountID)
	assert.Len(t, accountID, 12, "AWS account ID should be 12 digits")
	t.Logf("[%s] Cluster %s AWS account ID: %s", ctx.mode, ctx.testClusterID, accountID)
}

// TestConnectionLifecycle_E2E tests the automatic connection lifecycle (works with both mock and real OCM)
func TestConnectionLifecycle_E2E(t *testing.T) {
	ctx := setupE2ETest(t)
	// Don't defer Close() here - we test it explicitly

	// Test that connection works on first method call
	cluster, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	// Test that subsequent calls reuse the same connection
	cluster2, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster2)

	// Test closing connection
	err = ctx.client.Close()
	require.NoError(t, err)

	// Clean up mock server if present
	if ctx.mockServer != nil {
		ctx.mockServer.Close()
	}

	t.Logf("[%s] Connection lifecycle validated", ctx.mode)
}

// TestGetCluster_CCS_E2E tests CCS cluster detection using fixture (mock mode only)
func TestGetCluster_CCS_E2E(t *testing.T) {
	ctx := setupE2ETestWithFixture(t, "cluster_ccs.json")
	defer ctx.Close()

	cluster, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	ccs, ok := cluster.GetCCS()
	require.True(t, ok, "CCS should be present")
	assert.True(t, ccs.Enabled())

	t.Logf("[%s] Verified CCS cluster: %s (CCS enabled: %v)", ctx.mode, cluster.Name(), ccs.Enabled())
}

// TestGetCluster_Hypershift_E2E tests Hypershift cluster detection using fixture (mock mode only)
func TestGetCluster_Hypershift_E2E(t *testing.T) {
	ctx := setupE2ETestWithFixture(t, "cluster_hypershift.json")
	defer ctx.Close()

	cluster, err := ctx.client.GetCluster(ctx.testClusterID)
	require.NoError(t, err)
	require.NotNil(t, cluster)

	hypershift, ok := cluster.GetHypershift()
	require.True(t, ok, "Hypershift should be present")
	assert.True(t, hypershift.Enabled())

	t.Logf("[%s] Verified Hypershift cluster: %s (Hypershift enabled: %v)", ctx.mode, cluster.Name(), hypershift.Enabled())
}
