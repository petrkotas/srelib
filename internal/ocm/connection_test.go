package ocm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
