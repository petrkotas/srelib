# SDK Testing Infrastructure

This directory contains shared testing infrastructure for all SDK versions.

## Mock OCM Server

The `MockOCMServer` provides a test HTTP server that simulates the OCM API for testing purposes.

### Usage

```go
import sdktesting "github.com/petrkotas/srelib/sdk/testing"

func TestExample(t *testing.T) {
    // Create mock server
    mock := sdktesting.NewMockOCMServer(t)
    defer mock.Close()
    
    // Load fixtures
    clusterID, err := mock.LoadClusterFromFixture("cluster_osd.json")
    require.NoError(t, err)
    
    // Configure client to use mock server
    t.Setenv("OCM_URL", mock.Server.URL)
    t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
    
    // Run your tests...
}
```

### Fixtures

Fixtures are JSON files stored in `../testdata/fixtures/` (relative to the SDK root).

Available fixture loaders:

- `LoadClusterFromFixture(filename)` - Load cluster data
- `LoadSubscriptionFromFixture(filename)` - Load subscription data
- `LoadOrganizationFromFixture(filename)` - Load organization data

### Build Tags

The mock server is only compiled when the `integration` or `e2e` build tags are specified:

```bash
go test -tags=e2e ./sdk/v1/...
```

## Testing Helpers

The package also provides shared testing utilities:

### Fixture Loading

```go
data := sdktesting.LoadFixture(t, "cluster_osd.json")
```

### Validation Helpers

```go
// Validate cluster object
sdktesting.AssertClusterValid(t, cluster)

// Validate subscription object
sdktesting.AssertSubscriptionValid(t, subscription)

// Validate organization object
sdktesting.AssertOrganizationValid(t, organization)
```

### Utility Functions

```go
// Get environment variable with fallback
ocmURL := sdktesting.GetEnvOrDefault("OCM_URL", "staging")
```
