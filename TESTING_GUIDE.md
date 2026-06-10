# Testing Guide for SRELib

This guide explains the testing strategy for the SRELib SDK, particularly for end-to-end testing with the OCM (OpenShift Cluster Manager) API.

## Testing Architecture

### Test Types

1. **Mock Tests** (`-tags=integration`)
   - Use HTTP mock server with real OCM API fixtures
   - No credentials required
   - Fast execution
   - Test all code paths without hitting real OCM

2. **E2E Tests** (`-tags=e2e`)
   - Hit real OCM staging/production APIs
   - Require valid OCM credentials
   - Validate against real clusters
   - Slower but verify actual API compatibility

## Mock Testing with OCM

### Problem Solved

The OCM SDK doesn't expose its HTTP client, making traditional mocking difficult. Our solution:

1. Create an HTTP test server that mimics OCM API
2. Use real captured OCM responses as fixtures
3. Point OCM SDK to the mock server via environment variables
4. Test the **entire stack** except the final HTTP call

### Architecture Flow

```
Test Code
   ↓
RPCClient (sdk/v1/client.go)
   ↓
RPC Server (sdk/v1/server.go)
   ↓
i1.Client (internal/i1/types.go)
   ↓
OCM functions (internal/ocm/*.go)
   ↓
OCM SDK (github.com/openshift-online/ocm-sdk-go)
   ↓
HTTP Request → Mock Server (returns real fixture data)
```

**Everything is tested except the actual network call to OCM!**

## Running Tests

### Mock Tests (No Credentials)

```bash
# Run all mock tests
go test -v ./sdk/v1 -tags=integration -run "Mock"

# Run specific mock test
go test -v ./sdk/v1 -tags=integration -run "TestGetCluster_MockOCM_CCS"
```

### E2E Tests (Requires Credentials)

```bash
# Set credentials
export OCM_URL=staging  # or production, integration
export OCM_TOKEN=your-ocm-token
export TEST_CLUSTER_ID=your-test-cluster-id

# Run E2E tests
go test -v ./sdk/v1 -tags=e2e -run "RealOCM"
```

## Writing New Mock Tests

### Basic Pattern

```go
func TestYourFeature_MockOCM(t *testing.T) {
    // Skip if running with real credentials
    if getEnvOrDefault("OCM_TOKEN", "") != "" {
        t.Skip("Skipping mock test when OCM_TOKEN is set")
    }

    // Create mock server
    mockServer := NewMockOCMServer(t)
    defer mockServer.Close()

    // Load fixtures
    clusterID, err := mockServer.LoadClusterFromFixture("cluster_osd.json")
    if err != nil {
        t.Fatalf("Failed to load fixture: %v", err)
    }

    // Configure environment to use mock
    t.Setenv("OCM_URL", mockServer.Server.URL)
    t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")

    // Create client (will use mock server)
    client := setupTestClient(t)
    defer client.Close()

    // Test your functionality
    cluster, err := client.GetCluster(clusterID)
    require.NoError(t, err)
    assert.Equal(t, clusterID, cluster.ID())
}
```

### Available Fixtures

See [sdk/v1/testdata/fixtures/README.md](sdk/v1/testdata/fixtures/README.md) for details.

- `cluster_osd.json` - OpenShift Dedicated cluster
- `cluster_ccs.json` - CCS (Customer Cloud Subscription) cluster
- `cluster_hypershift.json` - Hypershift/HCP cluster
- `cluster_rosa.json` - ROSA cluster
- `subscription_response.json` - Subscription data
- `organization_response.json` - Organization data

### Loading Custom Fixtures

```go
// Load cluster fixture
clusterID, err := mockServer.LoadClusterFromFixture("cluster_ccs.json")

// Load subscription fixture
subID, err := mockServer.LoadSubscriptionFromFixture("subscription_response.json")

// Load organization fixture
orgID, err := mockServer.LoadOrganizationFromFixture("organization_response.json")

// Or add raw JSON for custom scenarios
customJSON := json.RawMessage(`{"id": "custom-123", "name": "custom", ...}`)
mockServer.AddClusterRaw("custom-123", customJSON)
```

## Adding New Fixtures

### Capture from Real OCM

```bash
# Set credentials
export OCM_URL=staging
export OCM_TOKEN=your-token

# Capture cluster
ocm get /api/clusters_mgmt/v1/clusters/CLUSTER_ID > temp.json

# Sanitize (REQUIRED - see below)
# ... sanitization steps ...

# Save to fixtures
mv sanitized.json sdk/v1/testdata/fixtures/cluster_new.json
```

### ⚠️ MANDATORY: Sanitize Fixtures

**NEVER commit fixtures with real customer data!**

Before adding any fixture:

1. Replace all real cluster IDs → `test-cluster-123`
2. Replace all AWS account IDs → `123456789012`
3. Replace all customer names → generic test names
4. Replace all real domains → `.test.t1.openshiftapps.com`
5. Remove any tokens, credentials, ARNs with real account IDs
6. Add `_metadata.sanitized: true`

See [sdk/v1/testdata/fixtures/README.md](sdk/v1/testdata/fixtures/README.md) for complete sanitization guide.

## Mock Server Features

### Supported Endpoints

- `GET /api/clusters_mgmt/v1/clusters` - List clusters (with search)
- `GET /api/clusters_mgmt/v1/clusters/{id}` - Get cluster by ID
- `GET /api/accounts_mgmt/v1/subscriptions` - List subscriptions (with search)
- `GET /api/accounts_mgmt/v1/organizations/{id}` - Get organization
- Authentication endpoints (returns dummy tokens)

### Search Syntax Support

The mock server parses OCM search syntax:

```go
// OCM search format: "id = 'value' or name = 'value'"
// Works in tests automatically
cluster, err := client.GetCluster("test-cluster-234")  // by ID
cluster, err := client.GetCluster("test-cluster-234s") // by name
```

### Indexing

Fixtures are automatically indexed by:
- Cluster: `id`, `name`, `external_id`
- Subscription: `id`, `cluster_id`, `display_name`
- Organization: `id`

## Benefits of This Approach

✅ **True E2E coverage** - Tests all your code including OCM integration layer
✅ **No credentials needed** - Mock tests run without OCM access
✅ **Fast** - No network calls to real APIs
✅ **Realistic** - Uses real OCM API response formats
✅ **Maintainable** - Easy to add new test scenarios
✅ **Safe** - Can't accidentally affect real clusters

## Best Practices

1. **Always sanitize fixtures** before committing
2. **Use mock tests for CI/CD** - fast and reliable
3. **Use E2E tests occasionally** - verify against real API changes
4. **Keep fixtures up to date** - when OCM API changes
5. **Document special fixtures** - if they test edge cases

## Troubleshooting

### Mock test fails with "invalid OCM_URL"

Make sure your `internal/ocm/connection.go` allows full URLs:

```go
// Should accept http:// and https:// URLs for testing
if strings.HasPrefix(urlEnv, "http://") || strings.HasPrefix(urlEnv, "https://") {
    ocmApiOverride = urlEnv
}
```

### Mock test fails with "Not logged in"

Ensure you're setting the dummy JWT token:

```go
t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")
```

### Fixture not found

Check the fixture path:
- Fixtures must be in `sdk/v1/testdata/fixtures/`
- Use the filename only, not the full path
- File must be valid JSON

## Future Improvements

- [ ] Add fixtures for more cluster types (GCP, Azure, etc.)
- [ ] Add fixtures for management cluster relationships
- [ ] Add fixtures for error scenarios
- [ ] Add helper to generate sanitized fixtures automatically
- [ ] Add fixture validation in CI to ensure sanitization

## References

- [Mock Server Implementation](sdk/v1/ocm_mock_server_test.go)
- [Fixture Documentation](sdk/v1/testdata/fixtures/README.md)
- [Connection Code](internal/ocm/connection.go)
- [E2E Tests](sdk/v1/client_e2e_test.go)
- [Integration Tests](sdk/v1/client_integration_test.go)
