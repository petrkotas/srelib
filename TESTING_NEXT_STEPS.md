# Testing Infrastructure - Next Steps

This document outlines recommended next steps and future improvements for the srelib testing infrastructure.

## Immediate Next Steps

### 1. Capture Real Test Fixtures (Priority: HIGH)

The current test fixtures in `sdk/v1/testdata/fixtures/` are synthesized based on code analysis and may not match actual OCM API responses.

**Action Items:**

#### Get Real Cluster Response
```bash
# Using OCM CLI (easiest method)
export OCM_URL=staging
export CLUSTER_ID="<your-test-cluster-id>"

# Capture cluster JSON
ocm get /api/clusters_mgmt/v1/clusters/$CLUSTER_ID | jq '.' > cluster_response.json

# For CCS cluster
ocm get /api/clusters_mgmt/v1/clusters/$CCS_CLUSTER_ID | jq '.' > cluster_ccs.json

# For Hypershift cluster
ocm get /api/clusters_mgmt/v1/clusters/$HCP_CLUSTER_ID | jq '.' > cluster_hypershift.json
```

#### Get Real Subscription Response
```bash
# Get subscription ID from cluster
SUB_ID=$(ocm get /api/clusters_mgmt/v1/clusters/$CLUSTER_ID | jq -r '.subscription.id')

# Capture subscription JSON
ocm get /api/accounts_mgmt/v1/subscriptions/$SUB_ID | jq '.' > subscription_response.json
```

#### Get Real Organization Response
```bash
# Get org ID from subscription
ORG_ID=$(ocm get /api/accounts_mgmt/v1/subscriptions/$SUB_ID | jq -r '.organization_id')

# Capture organization JSON
ocm get /api/accounts_mgmt/v1/organizations/$ORG_ID | jq '.' > organization_response.json
```

#### Sanitize Sensitive Data
```bash
cd sdk/v1/testdata/fixtures/

# Replace real IDs with test IDs
sed -i '' 's/real-cluster-id-here/test-cluster-123/g' *.json
sed -i '' 's/real-subscription-id-here/test-subscription-456/g' *.json
sed -i '' 's/real-org-id-here/test-org-999/g' *.json

# Remove sensitive ARNs and account IDs
jq 'del(.properties.rosa_creator_arn)' cluster_response.json > tmp.json && mv tmp.json cluster_response.json

# Replace AWS account IDs
sed -i '' 's/123456789012/000000000000/g' *.json
```

#### Document Fixture Source
Add a header comment to each fixture:
```json
{
  "_metadata": {
    "source": "OCM Staging API",
    "captured_date": "2024-06-08",
    "cluster_type": "standard|ccs|hypershift",
    "sanitized": true
  },
  "kind": "Cluster",
  ...
}
```

### 2. Verify E2E Tests Against Real OCM (Priority: HIGH)

**Action Items:**

1. Obtain OCM staging credentials:
   ```bash
   ocm login --url staging
   export OCM_TOKEN=$(ocm token)
   ```

2. Identify a stable test cluster in staging:
   ```bash
   # List available clusters
   ocm list clusters --url staging
   
   # Pick a long-lived test cluster
   export TEST_CLUSTER_ID="<stable-cluster-id>"
   ```

3. Run E2E tests:
   ```bash
   make test-e2e
   ```

4. Review test output:
   - Document any failing tests
   - Note permission issues (some tests may require elevated privileges)
   - Verify test coverage of different cluster types (standard, CCS, Hypershift)

5. Create a test cluster inventory document:
   ```bash
   # Document test resources
   cat > docs/test-clusters.md <<EOF
   # Test Cluster Inventory
   
   ## Staging Environment
   
   - **Standard Cluster**: cluster-id-xxx (created: 2024-01-15)
   - **CCS Cluster**: cluster-id-yyy (created: 2024-01-20)
   - **Hypershift Cluster**: cluster-id-zzz (created: 2024-02-01)
   
   ## Access
   
   OCM URL: staging
   Required permissions: cluster.read, subscription.read
   EOF
   ```

### 3. Set Up Continuous Testing (Priority: MEDIUM)

**If using GitHub Actions (future):**

Create `.github/workflows/test.yml`:
```yaml
name: Tests

on:
  push:
    branches: [main]
  pull_request:

jobs:
  unit-integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - run: make test

  e2e-staging:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: Run E2E tests
        env:
          OCM_TOKEN: ${{ secrets.OCM_STAGING_TOKEN }}
          TEST_CLUSTER_ID: ${{ secrets.TEST_CLUSTER_ID }}
        run: make test-e2e
```

**GitHub Secrets to Configure:**
- `OCM_STAGING_TOKEN` - Service account token for OCM staging
- `TEST_CLUSTER_ID` - Long-lived test cluster ID

**If using other CI (Jenkins, GitLab, etc.):**
- Create similar configuration
- Use secret management for credentials
- Run E2E tests only on main/master branch

## Future Improvements

### 1. Better Mocking Strategy (Priority: MEDIUM)

**Problem:** The OCM SDK doesn't expose HTTP client injection, making httptest-based mocking difficult.

**Potential Solutions:**

#### Option A: Wrapper Interface (Recommended)
Create an abstraction layer:

```go
// internal/ocm/interface.go
package ocm

type OCMClient interface {
    GetCluster(conn *sdk.Connection, key string) (*cmv1.Cluster, error)
    GetSubscription(conn *sdk.Connection, key string) (*amsv1.Subscription, error)
    // ... other methods
}

type DefaultOCMClient struct{}

func (c *DefaultOCMClient) GetCluster(conn *sdk.Connection, key string) (*cmv1.Cluster, error) {
    // Current implementation
}

// internal/ocm/mock.go (for tests)
type MockOCMClient struct {
    mock.Mock
}

func (m *MockOCMClient) GetCluster(conn *sdk.Connection, key string) (*cmv1.Cluster, error) {
    args := m.Called(conn, key)
    return args.Get(0).(*cmv1.Cluster), args.Error(1)
}
```

Then update `i1.Client` to use the interface:
```go
type Client struct {
    Logger    hclog.Logger
    ocmConn   *sdk.Connection
    ocmClient ocm.OCMClient  // <- injectable
}
```

**Effort:** Medium (requires refactoring)
**Benefit:** Full mock-based testing without OCM credentials

#### Option B: VCR Recording
Use go-vcr to record/replay real OCM interactions:

```go
// Record once with credentials
r, _ := recorder.New("testdata/vcr/get_cluster")
// ... make real API calls, recorded to cassette

// Replay in tests (no credentials)
r, _ := recorder.New("testdata/vcr/get_cluster")
r.SetMode(recorder.ModeReplayOnly)
```

**Effort:** Low-Medium (depends on OCM SDK HTTP accessibility)
**Benefit:** Tests against real API structure without mocking

### 2. Increase Test Coverage (Priority: MEDIUM)

**Current Coverage:**
- `internal/ocm`: 5.7%
- `sdk/v1`: 0.0% (only error path testing)

**Target Coverage:** 60-70% (industry standard for libraries)

**Action Items:**

1. Add unit tests for uncovered functions:
   - `internal/ocm/cluster.go` - All cluster operations
   - `internal/ocm/subscription.go` - All subscription operations
   - `internal/aws/account.go` - AWS account extraction

2. Test edge cases:
   - Empty responses
   - Malformed data
   - Network timeouts (if mockable)
   - Pagination handling

3. Generate coverage report:
   ```bash
   go test ./... -coverprofile=coverage.out
   go tool cover -html=coverage.out -o coverage.html
   ```

### 3. Performance Testing (Priority: LOW)

Add benchmarks for expensive operations:

```go
// sdk/v1/client_bench_test.go
func BenchmarkGetCluster(b *testing.B) {
    // Setup
    client := setupTestClient(b, "staging")
    defer client.CloseOCMConnection()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.GetCluster(testClusterID)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./sdk/v1
```

### 4. Test Data Builders (Priority: LOW)

Create builder pattern for test data:

```go
// sdk/v1/testdata/builders.go
type ClusterBuilder struct {
    cluster *cmv1.Cluster
}

func NewClusterBuilder() *ClusterBuilder {
    return &ClusterBuilder{
        cluster: cmv1.NewCluster().
            ID("test-cluster-123").
            Name("test-cluster").
            State(cmv1.ClusterStateReady),
    }
}

func (b *ClusterBuilder) WithCCS() *ClusterBuilder {
    b.cluster.CCS(cmv1.NewCCS().Enabled(true))
    return b
}

func (b *ClusterBuilder) WithHypershift() *ClusterBuilder {
    b.cluster.Hypershift(cmv1.NewHypershift().Enabled(true))
    return b
}

func (b *ClusterBuilder) Build() *cmv1.Cluster {
    return b.cluster
}
```

Usage in tests:
```go
cluster := NewClusterBuilder().
    WithCCS().
    Build()
```

### 5. Contract Testing (Priority: LOW)

If OCM API has OpenAPI/Swagger specs:

1. Download OCM API spec
2. Use tools like `go-swagger` or `prism` to validate responses
3. Run contract tests in CI:
   ```bash
   # Validate real responses match spec
   prism validate openapi.yaml --api-url staging
   ```

## Maintenance Tasks

### Ongoing

1. **Keep fixtures up to date** (Quarterly)
   - Re-capture fixtures when OCM API changes
   - Update tests if breaking changes occur

2. **Monitor E2E test stability** (Weekly)
   - Track flaky tests
   - Adjust timeouts if needed
   - Refresh test clusters if they get deleted

3. **Review test coverage** (Monthly)
   - Run coverage reports
   - Add tests for new functionality
   - Refactor duplicate test code

4. **Dependency updates** (Monthly)
   - Update testify: `go get -u github.com/stretchr/testify`
   - Update OCM SDK: `go get -u github.com/openshift-online/ocm-sdk-go`
   - Run tests after updates

### Before Major Releases

1. Run full test suite including E2E
2. Review and update TESTING.md documentation
3. Verify all test environments are accessible
4. Check test execution time (should be < 5min for unit+integration)

## Resources

- [Go Testing Best Practices](https://go.dev/wiki/TestComments)
- [testify Documentation](https://github.com/stretchr/testify)
- [OCM SDK Source](https://github.com/openshift-online/ocm-sdk-go)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Test Coverage Tools](https://go.dev/blog/cover)

## Questions / Decisions Needed

1. **Test Cluster Lifecycle**: Who manages test clusters in staging? How long should they live?
2. **CI/CD Platform**: GitHub Actions, Jenkins, GitLab, or other?
3. **Coverage Target**: What's the acceptable test coverage percentage?
4. **E2E Test Frequency**: How often should E2E tests run? (Every commit, daily, weekly?)
5. **Mocking Strategy**: Invest in wrapper interfaces now, or wait until more tests are needed?

## Tracking Progress

Create issues/tickets for:
- [ ] Capture real test fixtures from OCM staging
- [ ] Verify E2E tests with real credentials
- [ ] Set up CI/CD pipeline (if applicable)
- [ ] Decide on mocking strategy for future tests
- [ ] Increase test coverage to 60%+
- [ ] Document test cluster inventory
- [ ] Create service account for CI E2E tests
