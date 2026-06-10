# OCM API Response Fixtures

This directory contains **sanitized** OCM API responses captured from production/staging environments for use in testing.

⚠️ **IMPORTANT**: All fixtures in this directory MUST be sanitized to remove real customer data, credentials, and identifying information before being committed to the repository.

## Purpose

These fixtures enable true end-to-end testing by:
1. Providing real OCM API response formats
2. Allowing tests to run without OCM credentials  
3. Testing all OCM SDK code paths (`internal/ocm/`) against realistic data
4. Ensuring compatibility with actual OCM API responses

## Available Fixtures

### Cluster Fixtures

- **cluster_response.json** - Basic cluster response (minimal fields)
- **cluster_osd.json** - Full OpenShift Dedicated cluster
- **cluster_ccs.json** - Customer Cloud Subscription (CCS) cluster
- **cluster_hypershift.json** - Hypershift/HCP (Hosted Control Plane) cluster  
- **cluster_rosa.json** - Red Hat OpenShift Service on AWS cluster

### Subscription Fixtures

- **subscription_response.json** - Full subscription response with metrics

### Organization Fixtures

- **organization_response.json** - Organization details

## Metadata

Each fixture includes a `_metadata` section (when captured from real API):

```json
{
  "_metadata": {
    "source": "OCM Production API",
    "captured_date": "2026-06-09",
    "cluster_type": "ccs",
    "sanitized": true
  },
  ...
}
```

This metadata is for documentation only and is ignored by the OCM SDK.

## Using Fixtures in Tests

See `ocm_mock_server_test.go` for examples. The mock server loads fixtures and serves them via HTTP:

```go
mockServer := NewMockOCMServer(t)
defer mockServer.Close()

// Load a cluster fixture
clusterID, err := mockServer.LoadClusterFromFixture("cluster_osd.json")

// Point OCM SDK to mock server
t.Setenv("OCM_URL", mockServer.Server.URL)
t.Setenv("OCM_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjo5OTk5OTk5OTk5fQ.signature")

// Now all OCM calls use fixtures
client := setupTestClient(t)
cluster, err := client.GetCluster(clusterID)
```

## Capturing New Fixtures

To capture new fixtures from a real OCM environment:

```bash
# Set OCM credentials
export OCM_URL=staging  # or production, integration
export OCM_TOKEN=your-token-here

# Use curl or ocm CLI to fetch responses
ocm get /api/clusters_mgmt/v1/clusters/CLUSTER_ID | jq '.' > cluster_new.json

# Or via API
curl -H "Authorization: Bearer $OCM_TOKEN" \
  https://api.stage.openshift.com/api/clusters_mgmt/v1/clusters/CLUSTER_ID \
  | jq '.' > cluster_new.json
```

### ⚠️ MANDATORY: Sanitize Before Committing

**NEVER commit fixtures with real data!** Before adding any fixture to this directory, you MUST sanitize:

#### Data to Replace/Remove

1. **Customer Identifiers**
   - Real cluster IDs → Use test IDs like `test-cluster-123`
   - Real cluster names → Use generic names like `test-cluster-name`
   - External IDs → Generate random UUIDs or use test values
   - Organization IDs → Use `test-org-123`
   - Account IDs → Use `test-account-123`

2. **AWS/Cloud Credentials**
   - AWS Account IDs → Use `123456789012` (or other fake 12-digit numbers)
   - IAM Role ARNs → Replace account ID portions with fake IDs
   - Subnet IDs, VPC IDs → Use `subnet-12345678901234567` format
   - Any AWS resource identifiers

3. **Sensitive URLs and Endpoints**
   - Console URLs → Use `.test.t1.openshiftapps.com` domains
   - API URLs → Use test domains
   - OIDC endpoint URLs → Use test endpoints
   - Any customer-specific domains

4. **Authentication Tokens**
   - Any tokens, secrets, or credentials in responses
   - OAuth client IDs/secrets

5. **Personal/Customer Information**
   - Creator usernames/emails
   - Customer organization names → Use generic names
   - Any PII (personally identifiable information)

#### Sanitization Process

```bash
# 1. Capture the fixture
ocm get /api/clusters_mgmt/v1/clusters/REAL_ID > temp.json

# 2. Use jq or sed to replace sensitive data
jq '
  .id = "test-cluster-123" |
  .name = "test-cluster-123" |
  .external_id = "test-external-123" |
  .aws.billing_account_id = "123456789012" |
  # ... add more replacements as needed
' temp.json > cluster_sanitized.json

# 3. Add metadata marking it as sanitized
jq '. + {"_metadata": {"source": "OCM Production API", "captured_date": "2026-06-09", "sanitized": true}}' \
  cluster_sanitized.json > cluster_new.json

# 4. Manually review the file to ensure no real data remains
cat cluster_new.json | less

# 5. Delete temp file
rm temp.json cluster_sanitized.json
```

#### Verification Checklist

Before committing a new fixture, verify:

- [ ] No real cluster IDs or names
- [ ] No real AWS account IDs (all 12-digit numbers are fake)
- [ ] No real IAM role ARNs with customer account IDs  
- [ ] No real domain names (use `.test.t1.openshiftapps.com`)
- [ ] No real organization or customer names
- [ ] No tokens, credentials, or secrets
- [ ] `_metadata.sanitized: true` is present
- [ ] File has been manually reviewed line-by-line

**When in doubt, redact or use generic test values.**

## Fixture Format Requirements

All fixtures must:
1. Be valid JSON
2. Match the OCM API response schema exactly
3. Include required fields: `id`, `kind`, `href`
4. Have unique IDs to avoid conflicts in tests
5. **Be fully sanitized with no real customer data** (see sanitization section above)
6. Include `_metadata.sanitized: true` to indicate sanitization was performed

## Running Tests

```bash
# Run mock tests (no OCM credentials needed)
go test -v ./sdk/v1 -run "TestGetCluster_MockOCM" -tags=integration

# Run real E2E tests (requires OCM credentials)
export OCM_TOKEN=your-token
export TEST_CLUSTER_ID=your-cluster-id
go test -v ./sdk/v1 -run "RealOCM" -tags=e2e
```

## Maintenance

- Review fixtures periodically for OCM API changes
- Update fixtures when adding new test scenarios
- Keep fixture IDs consistent with related test code
- Document any special fixture characteristics in this README
