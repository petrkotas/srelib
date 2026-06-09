package v1

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-hclog"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/petrkotas/srelib/internal/i1"
)

// setupTestClient creates a client for testing
// For integration tests with mock servers, pass the mock server URL
// For E2E tests, pass empty string to use default OCM config
func setupTestClient(t *testing.T, ocmURL string) *i1.Client {
	logger := hclog.New(&hclog.LoggerOptions{
		Level: hclog.Error, // Quiet logging during tests
	})

	client, err := i1.NewClient(logger)
	require.NoError(t, err, "Failed to create test client")
	return client
}

// loadFixture loads a JSON fixture file from testdata/fixtures/
func loadFixture(t *testing.T, filename string) []byte {
	// Get the directory of this source file
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get caller information")

	dir := filepath.Dir(file)
	fixturePath := filepath.Join(dir, "testdata", "fixtures", filename)

	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "Failed to read fixture file: %s", filename)

	return data
}

// assertClusterValid checks common cluster validation rules
func assertClusterValid(t *testing.T, cluster *cmv1.Cluster) {
	assert.NotNil(t, cluster, "Cluster should not be nil")
	if cluster == nil {
		return
	}

	assert.NotEmpty(t, cluster.ID(), "Cluster ID should not be empty")
	assert.NotEmpty(t, cluster.Name(), "Cluster name should not be empty")
}

// assertSubscriptionValid checks subscription validation rules
func assertSubscriptionValid(t *testing.T, sub *amsv1.Subscription) {
	assert.NotNil(t, sub, "Subscription should not be nil")
	if sub == nil {
		return
	}

	assert.NotEmpty(t, sub.ID(), "Subscription ID should not be empty")
}

// assertOrganizationValid checks organization validation rules
func assertOrganizationValid(t *testing.T, org *amsv1.Organization) {
	assert.NotNil(t, org, "Organization should not be nil")
	if org == nil {
		return
	}

	assert.NotEmpty(t, org.ID(), "Organization ID should not be empty")
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
