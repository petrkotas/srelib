//go:build integration || e2e
// +build integration e2e

package testing

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadFixture loads a JSON fixture file from testdata/fixtures/
// Looks for fixtures relative to the sdk directory (one level up from testing/)
func LoadFixture(t *testing.T, filename string) []byte {
	// Get the directory of this source file
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get caller information")

	dir := filepath.Dir(file)
	fixturePath := filepath.Join(dir, "..", "testdata", "fixtures", filename)

	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "Failed to read fixture file: %s", filename)

	return data
}

// AssertClusterValid checks common cluster validation rules
func AssertClusterValid(t *testing.T, cluster *cmv1.Cluster) {
	assert.NotNil(t, cluster, "Cluster should not be nil")
	if cluster == nil {
		return
	}

	assert.NotEmpty(t, cluster.ID(), "Cluster ID should not be empty")
	assert.NotEmpty(t, cluster.Name(), "Cluster name should not be empty")
}

// AssertSubscriptionValid checks subscription validation rules
func AssertSubscriptionValid(t *testing.T, sub *amsv1.Subscription) {
	assert.NotNil(t, sub, "Subscription should not be nil")
	if sub == nil {
		return
	}

	assert.NotEmpty(t, sub.ID(), "Subscription ID should not be empty")
}

// AssertOrganizationValid checks organization validation rules
func AssertOrganizationValid(t *testing.T, org *amsv1.Organization) {
	assert.NotNil(t, org, "Organization should not be nil")
	if org == nil {
		return
	}

	assert.NotEmpty(t, org.ID(), "Organization ID should not be empty")
}

// GetEnvOrDefault returns environment variable value or default if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
