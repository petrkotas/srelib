//go:build integration || e2e
// +build integration e2e

package v1

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"

	"github.com/petrkotas/srelib/internal/i1"
)

// setupTestClient creates a v1 client for testing
// Uses environment variables OCM_URL and OCM_TOKEN to configure the connection.
// This bypasses the OCM config file requirement for testing purposes.
func setupTestClient(t *testing.T) *i1.Client {
	logger := hclog.New(&hclog.LoggerOptions{
		Level: hclog.Error, // Quiet logging during tests
	})

	client, err := i1.NewTestClient(logger)
	require.NoError(t, err, "Failed to create test client")
	return client
}
