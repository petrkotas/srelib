package ocm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOCMConfigLocation(t *testing.T) {
	// Test with OCM_CONFIG environment variable
	testPath := "/tmp/test-ocm.json"
	os.Setenv("OCM_CONFIG", testPath)
	defer os.Unsetenv("OCM_CONFIG")

	location, err := GetOCMConfigLocation()
	require.NoError(t, err, "GetOCMConfigLocation should not return an error")
	assert.Equal(t, testPath, location, "OCM config location should match OCM_CONFIG environment variable")
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
				assert.Error(t, err, "Expected error for input %s", tt.input)
			} else {
				require.NoError(t, err, "Unexpected error for input %s", tt.input)
				assert.Equal(t, tt.expected, result, "Expected %s, got %s", tt.expected, result)
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
		resolved, ok := urlAliases[alias]
		assert.True(t, ok, "Alias %s should be found in urlAliases map", alias)
		assert.Equal(t, expectedURL, resolved, "Alias %s should map to %s", alias, expectedURL)
	}
}
