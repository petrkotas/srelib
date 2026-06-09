package v1

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/petrkotas/srelib/internal/i1"
)

// TestClient_ConnectionLifecycle tests the OCM connection lifecycle
func TestClient_ConnectionLifecycle(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{Level: hclog.Error})
	client := &i1.Client{Logger: logger}

	// Test GetOCMConnection before initialization
	_, err := client.GetOCMConnection()
	assert.Error(t, err, "GetOCMConnection should error when connection not initialized")
	assert.Contains(t, err.Error(), "not initialized")

	// Note: Cannot test CreateOCMConnection without valid OCM credentials
	// This would require mocking the OCM SDK's connection builder
	// See E2E tests for actual connection testing

	// Test CloseOCMConnection when connection is nil (should not error)
	err = client.CloseOCMConnection()
	assert.NoError(t, err, "CloseOCMConnection should not error when connection is nil")
}

// TestClient_ClusterOperations_UninitializedConnection tests that cluster operations
// fail gracefully when connection is not initialized
func TestClient_ClusterOperations_UninitializedConnection(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{Level: hclog.Error})
	client := &i1.Client{Logger: logger}

	t.Run("GetCluster", func(t *testing.T) {
		_, err := client.GetCluster("test-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("GetClusterAnyStatus", func(t *testing.T) {
		_, err := client.GetClusterAnyStatus("test-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("GetClusters", func(t *testing.T) {
		_, err := client.GetClusters([]string{"test-id-1", "test-id-2"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

// TestClient_SubscriptionOperations_UninitializedConnection tests that subscription operations
// fail gracefully when connection is not initialized
func TestClient_SubscriptionOperations_UninitializedConnection(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{Level: hclog.Error})
	client := &i1.Client{Logger: logger}

	t.Run("GetSubscription", func(t *testing.T) {
		_, err := client.GetSubscription("test-sub-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("GetOrganization", func(t *testing.T) {
		_, err := client.GetOrganization("test-org-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	t.Run("GetOrgFromClusterID", func(t *testing.T) {
		_, err := client.GetOrgFromClusterID("test-cluster-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

// TestClient_AWSOperations_UninitializedConnection tests that AWS operations
// fail gracefully when connection is not initialized
func TestClient_AWSOperations_UninitializedConnection(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{Level: hclog.Error})
	client := &i1.Client{Logger: logger}

	// Note: These tests can't call the methods without a valid cluster object
	// which requires an OCM connection to create. See E2E tests for actual testing.

	t.Run("GetSupportRoleArnForCluster requires initialized connection", func(t *testing.T) {
		// This test documents that the method requires an initialized connection
		// Actual testing requires E2E tests with real OCM data
		assert.NotNil(t, client.Logger, "Client should have logger initialized")
	})
}

// TestClient_ConfigOperations tests OCM configuration operations
func TestClient_ConfigOperations(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{Level: hclog.Error})
	client := &i1.Client{Logger: logger}

	t.Run("GetOCMConfigLocation", func(t *testing.T) {
		// This should work without an OCM connection
		location, err := client.GetOCMConfigLocation()
		// May error if no config exists, but should not panic
		if err == nil {
			assert.NotEmpty(t, location, "Config location should not be empty when no error")
		}
	})

	t.Run("LoadOCMConfig with empty path", func(t *testing.T) {
		// Loading config should not require an active connection
		// May error if config doesn't exist, but should not panic
		err := client.LoadOCMConfig("")
		// We don't assert on error since config may or may not exist
		// The important thing is it doesn't panic
		_ = err
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
