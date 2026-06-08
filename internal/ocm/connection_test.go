package ocm

import (
	"os"
	"testing"
)

func TestGetOCMConfigLocation(t *testing.T) {
	// Test with OCM_CONFIG environment variable
	testPath := "/tmp/test-ocm.json"
	os.Setenv("OCM_CONFIG", testPath)
	defer os.Unsetenv("OCM_CONFIG")

	location, err := GetOCMConfigLocation()
	if err != nil {
		t.Fatalf("GetOCMConfigLocation() failed: %v", err)
	}

	if location != testPath {
		t.Errorf("Expected location %s, got %s", testPath, location)
	}
}

func TestValidateAndResolveOcmUrl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{"production alias", "production", productionURL, false},
		{"prod alias", "prod", productionURL, false},
		{"staging alias", "staging", stagingURL, false},
		{"integration alias", "integration", integrationURL, false},
		{"full production URL", productionURL, productionURL, false},
		{"invalid alias", "invalid", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateAndResolveOcmUrl(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %s, got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestURLAliases(t *testing.T) {
	expectedMappings := map[string]string{
		"production":    productionURL,
		"prod":          productionURL,
		"prd":           productionURL,
		"staging":       stagingURL,
		"stage":         stagingURL,
		"stg":           stagingURL,
		"integration":   integrationURL,
		"int":           integrationURL,
		"productiongov": productionGovURL,
		"prodgov":       productionGovURL,
		"prdgov":        productionGovURL,
	}

	for alias, expectedURL := range expectedMappings {
		if resolved, ok := urlAliases[alias]; !ok {
			t.Errorf("Alias %s not found in urlAliases map", alias)
		} else if resolved != expectedURL {
			t.Errorf("Alias %s maps to %s, expected %s", alias, resolved, expectedURL)
		}
	}
}
