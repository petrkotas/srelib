package ocm

import (
	"fmt"
	"os"
	"strings"

	sdk "github.com/openshift-online/ocm-sdk-go"
)

// CreateTestConnection creates a connection to OCM specifically for testing purposes.
// This function bypasses the OCM config file requirement and allows direct URL + token configuration.
//
// Environment variables:
// - OCM_URL: URL override - can be an alias ('staging', 'production', etc.) or a full URL (including mock server URLs)
// - OCM_TOKEN: Authentication token (required for test connections)
//
// This function should ONLY be used in test code. Production code should use CreateConnection() instead.
func CreateTestConnection() (*sdk.Connection, error) {
	urlEnv := os.Getenv("OCM_URL")
	tokenEnv := os.Getenv("OCM_TOKEN")

	if tokenEnv == "" {
		return nil, fmt.Errorf("OCM_TOKEN is required for test connections")
	}

	var ocmApiURL string
	if urlEnv != "" {
		// Check if it's an alias first
		gatewayURL, ok := urlAliases[urlEnv]
		if ok {
			ocmApiURL = gatewayURL
		} else if strings.HasPrefix(urlEnv, "http://") || strings.HasPrefix(urlEnv, "https://") {
			// Allow full URLs (for testing with mock servers)
			ocmApiURL = urlEnv
		} else {
			return nil, fmt.Errorf("invalid OCM_URL found: %s\nValid URL aliases are: 'production', 'staging', 'integration', 'productiongov', 'staginggov', 'integrationgov' or a full URL", urlEnv)
		}
	} else {
		// Default to production if no URL specified
		ocmApiURL = productionURL
	}

	agentString := fmt.Sprintf("srelib-%s", Version)

	// Create a simple connection using URL and token directly
	return sdk.NewConnectionBuilder().
		URL(ocmApiURL).
		TokenURL(ocmApiURL).
		Tokens(tokenEnv).
		Agent(agentString).
		Build()
}
