//go:build integration || e2e
// +build integration e2e

package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file contains only the MockOCMServer infrastructure.
// For actual tests using the mock server, see client_e2e_test.go

// MockOCMServer provides a mock OCM API server for testing with real OCM response fixtures
type MockOCMServer struct {
	Server        *httptest.Server
	Clusters      map[string]json.RawMessage
	Subscriptions map[string]json.RawMessage
	Organizations map[string]json.RawMessage
	fixturesPath  string
}

// NewMockOCMServer creates a new mock OCM API server
func NewMockOCMServer(t *testing.T) *MockOCMServer {
	// Find fixtures directory - go up to sdk level
	fixturesPath := filepath.Join("..", "testdata", "fixtures")

	mock := &MockOCMServer{
		Clusters:      make(map[string]json.RawMessage),
		Subscriptions: make(map[string]json.RawMessage),
		Organizations: make(map[string]json.RawMessage),
		fixturesPath:  fixturesPath,
	}

	// Create a custom handler that handles all requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request for debugging
		// t.Logf("Mock server received: %s %s", r.Method, r.URL.Path)

		// Handle specific API endpoints
		switch {
		case r.URL.Path == "/api/clusters_mgmt/v1/clusters" && r.Method == http.MethodGet:
			mock.handleClustersListRequest(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/clusters_mgmt/v1/clusters/") && r.Method == http.MethodGet:
			mock.handleClusterByIDRequest(w, r)
		case r.URL.Path == "/api/accounts_mgmt/v1/subscriptions" && r.Method == http.MethodGet:
			mock.handleSubscriptionsRequest(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/accounts_mgmt/v1/organizations/") && r.Method == http.MethodGet:
			mock.handleOrganizationsRequest(w, r)
		default:
			// Return JSON for anything else (like auth token requests)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "mock-access-token",
				"refresh_token": "mock-refresh-token",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
		}
	})

	mock.Server = httptest.NewServer(handler)
	return mock
}

// Close shuts down the mock server
func (m *MockOCMServer) Close() {
	m.Server.Close()
}

// LoadClusterFromFixture loads a cluster fixture file and adds it to the mock server
// Returns the cluster ID for convenience
func (m *MockOCMServer) LoadClusterFromFixture(fixtureName string) (string, error) {
	fixturePath := filepath.Join(m.fixturesPath, fixtureName)
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return "", fmt.Errorf("failed to read fixture %s: %w", fixtureName, err)
	}

	// Parse to get the cluster ID
	var clusterData map[string]interface{}
	if err := json.Unmarshal(data, &clusterData); err != nil {
		return "", fmt.Errorf("failed to parse fixture %s: %w", fixtureName, err)
	}

	clusterID, ok := clusterData["id"].(string)
	if !ok {
		return "", fmt.Errorf("fixture %s missing 'id' field", fixtureName)
	}

	// Store raw JSON for exact responses
	m.Clusters[clusterID] = data

	// Also index by name and external_id for search support
	if name, ok := clusterData["name"].(string); ok && name != "" {
		m.Clusters[name] = data
	}
	if externalID, ok := clusterData["external_id"].(string); ok && externalID != "" {
		m.Clusters[externalID] = data
	}

	return clusterID, nil
}

// LoadSubscriptionFromFixture loads a subscription fixture file
func (m *MockOCMServer) LoadSubscriptionFromFixture(fixtureName string) (string, error) {
	fixturePath := filepath.Join(m.fixturesPath, fixtureName)
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return "", fmt.Errorf("failed to read fixture %s: %w", fixtureName, err)
	}

	var subData map[string]interface{}
	if err := json.Unmarshal(data, &subData); err != nil {
		return "", fmt.Errorf("failed to parse fixture %s: %w", fixtureName, err)
	}

	subID, ok := subData["id"].(string)
	if !ok {
		return "", fmt.Errorf("fixture %s missing 'id' field", fixtureName)
	}

	m.Subscriptions[subID] = data

	// Index by cluster_id and display_name for search
	if clusterID, ok := subData["cluster_id"].(string); ok && clusterID != "" {
		m.Subscriptions[clusterID] = data
	}
	if displayName, ok := subData["display_name"].(string); ok && displayName != "" {
		m.Subscriptions[displayName] = data
	}

	return subID, nil
}

// LoadOrganizationFromFixture loads an organization fixture file
func (m *MockOCMServer) LoadOrganizationFromFixture(fixtureName string) (string, error) {
	fixturePath := filepath.Join(m.fixturesPath, fixtureName)
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return "", fmt.Errorf("failed to read fixture %s: %w", fixtureName, err)
	}

	var orgData map[string]interface{}
	if err := json.Unmarshal(data, &orgData); err != nil {
		return "", fmt.Errorf("failed to parse fixture %s: %w", fixtureName, err)
	}

	orgID, ok := orgData["id"].(string)
	if !ok {
		return "", fmt.Errorf("fixture %s missing 'id' field", fixtureName)
	}

	m.Organizations[orgID] = data
	return orgID, nil
}

// AddClusterRaw adds a cluster using raw JSON (for custom test scenarios)
func (m *MockOCMServer) AddClusterRaw(id string, rawJSON json.RawMessage) {
	m.Clusters[id] = rawJSON
}

func (m *MockOCMServer) handleClustersListRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse search query - OCM uses format like: "id = 'xyz' or name = 'xyz'"
	search := r.URL.Query().Get("search")

	var matchedClusters []json.RawMessage
	seen := make(map[string]bool) // Deduplicate since we index by id, name, and external_id

	if search != "" {
		// Extract search terms from OCM search syntax
		// Format: "id = 'value' or name = 'value' or external_id = 'value'"
		searchTerms := extractSearchTerms(search)

		for _, term := range searchTerms {
			if clusterJSON, exists := m.Clusters[term]; exists {
				// Parse to get unique ID for deduplication
				var clusterData map[string]interface{}
				if err := json.Unmarshal(clusterJSON, &clusterData); err == nil {
					if id, ok := clusterData["id"].(string); ok {
						if !seen[id] {
							seen[id] = true
							matchedClusters = append(matchedClusters, clusterJSON)
						}
					}
				}
			}
		}
	} else {
		// No search - return all unique clusters
		for key, clusterJSON := range m.Clusters {
			if !seen[key] {
				var clusterData map[string]interface{}
				if err := json.Unmarshal(clusterJSON, &clusterData); err == nil {
					if id, ok := clusterData["id"].(string); ok {
						if !seen[id] {
							seen[id] = true
							matchedClusters = append(matchedClusters, clusterJSON)
						}
					}
				}
			}
		}
	}

	// Build OCM-style list response
	response := map[string]interface{}{
		"kind":  "ClusterList",
		"total": len(matchedClusters),
		"size":  len(matchedClusters),
		"page":  1,
		"items": matchedClusters,
	}

	json.NewEncoder(w).Encode(response)
}

// extractSearchTerms parses OCM search syntax and extracts the actual search values
// Example: "id = 'test-123' or name = 'my-cluster'" -> ["test-123", "my-cluster"]
func extractSearchTerms(search string) []string {
	var terms []string

	// Simple parser for OCM search syntax
	// Look for quoted strings
	parts := strings.Split(search, "'")
	for i := 1; i < len(parts); i += 2 {
		term := strings.TrimSpace(parts[i])
		if term != "" {
			terms = append(terms, term)
		}
	}

	return terms
}

func (m *MockOCMServer) handleClusterByIDRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract cluster ID from path: /api/clusters_mgmt/v1/clusters/{id}
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	clusterID := parts[len(parts)-1]

	clusterJSON, exists := m.Clusters[clusterID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind":   "Error",
			"reason": "Cluster not found",
			"code":   "CLUSTERS-MGMT-404",
		})
		return
	}

	// Return raw JSON response from fixture
	w.Write(clusterJSON)
}

func (m *MockOCMServer) handleSubscriptionsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse search parameter
	search := r.URL.Query().Get("search")
	parameter := r.URL.Query().Get("parameter")

	var matchedSubscriptions []json.RawMessage
	seen := make(map[string]bool)

	if search != "" || parameter != "" {
		searchTerms := extractSearchTerms(search)
		if parameter != "" {
			searchTerms = append(searchTerms, extractSearchTerms(parameter)...)
		}

		for _, term := range searchTerms {
			if subJSON, exists := m.Subscriptions[term]; exists {
				var subData map[string]interface{}
				if err := json.Unmarshal(subJSON, &subData); err == nil {
					if id, ok := subData["id"].(string); ok {
						if !seen[id] {
							seen[id] = true
							matchedSubscriptions = append(matchedSubscriptions, subJSON)
						}
					}
				}
			}
		}
	} else {
		for key, subJSON := range m.Subscriptions {
			if !seen[key] {
				var subData map[string]interface{}
				if err := json.Unmarshal(subJSON, &subData); err == nil {
					if id, ok := subData["id"].(string); ok {
						if !seen[id] {
							seen[id] = true
							matchedSubscriptions = append(matchedSubscriptions, subJSON)
						}
					}
				}
			}
		}
	}

	response := map[string]interface{}{
		"kind":  "SubscriptionList",
		"total": len(matchedSubscriptions),
		"size":  len(matchedSubscriptions),
		"page":  1,
		"items": matchedSubscriptions,
	}

	json.NewEncoder(w).Encode(response)
}

func (m *MockOCMServer) handleOrganizationsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract org ID from path: /api/accounts_mgmt/v1/organizations/{id}
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	orgID := parts[len(parts)-1]

	orgJSON, exists := m.Organizations[orgID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind":   "Error",
			"reason": "Organization not found",
			"code":   "ACCT-MGMT-404",
		})
		return
	}

	// Return raw JSON response from fixture
	w.Write(orgJSON)
}
