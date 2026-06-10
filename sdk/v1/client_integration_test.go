package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_ConnectionLifecycle tests the automatic OCM connection lifecycle
func TestClient_ConnectionLifecycle(t *testing.T) {
	client := setupTestClient(t)

	// Test that Close works even without any operations
	err := client.Close()
	assert.NoError(t, err, "Close should not error even if connection was never initialized")

	// Test Close can be called multiple times
	err = client.Close()
	assert.NoError(t, err, "Close should be safe to call multiple times")
}

// TestClient_ClusterOperations_MissingCredentials tests that cluster operations
// fail gracefully when OCM credentials are not available
func TestClient_ClusterOperations_MissingCredentials(t *testing.T) {
	// Clear OCM-related env vars to simulate missing credentials
	t.Setenv("OCM_TOKEN", "")
	t.Setenv("OCM_CONFIG", "/nonexistent/path/ocm.json")

	client := setupTestClient(t)

	t.Run("GetCluster fails with missing credentials", func(t *testing.T) {
		_, err := client.GetCluster("test-id")
		require.Error(t, err, "Should fail when credentials are missing")
		assert.Contains(t, err.Error(), "failed to create OCM connection")
	})

	t.Run("GetClusterAnyStatus fails with missing credentials", func(t *testing.T) {
		_, err := client.GetClusterAnyStatus("test-id")
		require.Error(t, err, "Should fail when credentials are missing")
		assert.Contains(t, err.Error(), "failed to create OCM connection")
	})

	t.Run("GetClusters fails with missing credentials", func(t *testing.T) {
		_, err := client.GetClusters([]string{"test-id-1", "test-id-2"})
		require.Error(t, err, "Should fail when credentials are missing")
		assert.Contains(t, err.Error(), "failed to create OCM connection")
	})
}

// TestClient_SubscriptionOperations_MissingCredentials tests that subscription operations
// fail gracefully when OCM credentials are not available
func TestClient_SubscriptionOperations_MissingCredentials(t *testing.T) {
	// Clear OCM-related env vars to simulate missing credentials
	t.Setenv("OCM_TOKEN", "")
	t.Setenv("OCM_CONFIG", "/nonexistent/path/ocm.json")

	client := setupTestClient(t)

	t.Run("GetSubscription fails with missing credentials", func(t *testing.T) {
		_, err := client.GetSubscription("test-sub-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create OCM connection")
	})

	t.Run("GetOrganization fails with missing credentials", func(t *testing.T) {
		_, err := client.GetOrganization("test-org-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create OCM connection")
	})

}

// TestClient_AWSOperations_MissingCredentials tests that AWS operations
// fail gracefully when OCM credentials are not available
func TestClient_AWSOperations_MissingCredentials(t *testing.T) {
	// Note: These tests can't call the methods without a valid cluster object
	// which requires an OCM connection to create. See E2E tests for actual testing.

	t.Run("AWS operations require valid OCM connection", func(t *testing.T) {
		// This test documents that AWS methods require an initialized connection
		// Actual testing requires E2E tests with real OCM data
		client := setupTestClient(t)
		assert.NotNil(t, client, "Client should be created")
	})
}

// NOTE: Full integration testing of GetCluster, GetSubscription, etc. with mocked
// OCM API responses is not feasible with the current architecture because:
// 1. The OCM SDK (ocm-sdk-go) manages its own HTTP client internally
// 2. The connection builder doesn't expose an option to inject a custom HTTP client
// 3. We cannot intercept HTTP requests without modifying the OCM SDK
//
// For comprehensive testing of these methods, see:
// - E2E tests (client_e2e_test.go) - tests against real OCM staging environment
// - Unit tests in internal/ocm/*_test.go - tests utility functions with mocked data
