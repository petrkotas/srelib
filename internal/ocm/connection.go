package ocm

import (
	"fmt"
	"os"

	ocmConfig "github.com/openshift-online/ocm-common/pkg/ocm/config"
	ocmConnBuilder "github.com/openshift-online/ocm-common/pkg/ocm/connection-builder"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

const (
	productionURL     = "https://api.openshift.com"
	stagingURL        = "https://api.stage.openshift.com"
	integrationURL    = "https://api.integration.openshift.com"
	productionGovURL  = "https://api-admin.openshiftusgov.com"
	integrationGovURL = "https://api-admin.int.openshiftusgov.com"
	stagingGovURL     = "https://api-admin.stage.openshiftusgov.com"
)

var urlAliases = map[string]string{
	"production":      productionURL,
	"prod":            productionURL,
	"prd":             productionURL,
	productionURL:     productionURL,
	"staging":         stagingURL,
	"stage":           stagingURL,
	"stg":             stagingURL,
	stagingURL:        stagingURL,
	"integration":     integrationURL,
	"int":             integrationURL,
	integrationURL:    integrationURL,
	"productiongov":   productionGovURL,
	"prodgov":         productionGovURL,
	"prdgov":          productionGovURL,
	productionGovURL:  productionGovURL,
	"integrationgov":  integrationGovURL,
	"intgov":          integrationGovURL,
	integrationGovURL: integrationGovURL,
	"staginggov":      stagingGovURL,
	"stagegov":        stagingGovURL,
	stagingGovURL:     stagingGovURL,
}

// Version is set at build time
var Version = "dev"

// CreateConnection creates a connection to OCM using the OCM config file.
// To use OCM url override set OCM_URL environment variable to one of the following values:
// 'production', 'staging', 'integration', 'productiongov', 'staginggov', 'integrationgov' or the full URL of the OCM API.
//
// This function is for production use and requires a valid OCM config file (~/.config/ocm/ocm.json).
// For testing purposes, use CreateTestConnection() instead.
func CreateConnection() (*sdk.Connection, error) {
	urlEnv := os.Getenv("OCM_URL")
	var ocmApiOverride string

	if urlEnv != "" {
		// Check if it's an alias first
		gatewayURL, ok := urlAliases[urlEnv]
		if !ok {
			return nil, fmt.Errorf("invalid OCM_URL found: %s\nValid URL aliases are: 'production', 'staging', 'integration', 'productiongov', 'staginggov', 'integrationgov' or a full URL", urlEnv)
		}

		ocmApiOverride = gatewayURL
	}

	agentString := fmt.Sprintf("srelib-%s", Version)

	// Load from OCM config file
	config, err := ocmConfig.Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load OCM config: %w", err)
	}

	connBuilder := ocmConnBuilder.NewConnection().Config(config).AsAgent(agentString)

	if ocmApiOverride != "" {
		connBuilder.WithApiUrl(ocmApiOverride)
	}

	return connBuilder.Build()
}
