package ocm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdk "github.com/openshift-online/ocm-sdk-go"
	ocmConfig "github.com/openshift-online/ocm-common/pkg/ocm/config"
	ocmConnBuilder "github.com/openshift-online/ocm-common/pkg/ocm/connection-builder"
)

const (
	productionURL         = "https://api.openshift.com"
	stagingURL            = "https://api.stage.openshift.com"
	integrationURL        = "https://api.integration.openshift.com"
	productionGovURL      = "https://api-admin.openshiftusgov.com"
	integrationGovURL     = "https://api-admin.int.openshiftusgov.com"
	stagingGovURL         = "https://api-admin.stage.openshiftusgov.com"
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

// CreateConnection creates a connection to OCM using default configuration
func CreateConnection() (*sdk.Connection, error) {
	urlEnv := os.Getenv("OCM_URL")
	var ocmApiOverride string
	if urlEnv != "" {
		gatewayURL, ok := urlAliases[urlEnv]
		if !ok {
			return nil, fmt.Errorf("invalid OCM_URL found: %s\nValid URL aliases are: 'production', 'staging', 'integration'", urlEnv)
		}

		ocmApiOverride = gatewayURL
	}

	config, err := ocmConfig.Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load OCM config: %w", err)
	}

	agentString := fmt.Sprintf("srelib-%s", Version)

	connBuilder := ocmConnBuilder.NewConnection().Config(config).AsAgent(agentString)

	if ocmApiOverride != "" {
		connBuilder.WithApiUrl(ocmApiOverride)
	}

	return connBuilder.Build()
}

// CreateConnectionWithUrl creates a connection to OCM with a specific URL
func CreateConnectionWithUrl(ocmUrl string) (*sdk.Connection, error) {
	ocmApiUrl, err := ValidateAndResolveOcmUrl(ocmUrl)
	if err != nil {
		return nil, err
	}

	config, err := ocmConfig.Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load OCM config: %w", err)
	}

	agentString := fmt.Sprintf("srelib-%s", Version)

	connBuilder := ocmConnBuilder.NewConnection().Config(config).AsAgent(agentString)

	if connBuilder == nil {
		return nil, fmt.Errorf("ocm connection builder returned nil")
	}
	connBuilder.WithApiUrl(ocmApiUrl)

	return connBuilder.Build()
}

// ValidateAndResolveOcmUrl validates an OCM URL or alias and resolves it to a full URL.
func ValidateAndResolveOcmUrl(ocmUrl string) (string, error) {
	if len(ocmUrl) <= 0 {
		return "", fmt.Errorf("empty OCM URL")
	}

	resolvedUrl, ok := urlAliases[ocmUrl]
	if !ok {
		return "", fmt.Errorf("invalid OCM_URL found: %s\nValid URL aliases are: 'production', 'staging', 'integration'", ocmUrl)
	}
	return resolvedUrl, nil
}

// GetOCMConfigLocation finds the OCM configuration file and returns the path to it
func GetOCMConfigLocation() (string, error) {
	if ocmconfig := os.Getenv("OCM_CONFIG"); ocmconfig != "" {
		return ocmconfig, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".ocm.json")

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		path = filepath.Join(configDir, "/ocm/ocm.json")
	}

	return path, nil
}

// LoadOCMConfig loads the OCM configuration file
func LoadOCMConfig() (*ocmConfig.Config, error) {
	file, err := GetOCMConfigLocation()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		return &ocmConfig.Config{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("can't check if config file '%s' exists: %w", file, err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("can't read config file '%s': %w", file, err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	cfg := &ocmConfig.Config{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't parse config file '%s': %w", file, err)
	}

	return cfg, nil
}

// LoadOCMConfigFromPath loads the OCM configuration from a specific file path
func LoadOCMConfigFromPath(filePath string) (*ocmConfig.Config, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return &ocmConfig.Config{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("can't check if config file '%s' exists: %w", filePath, err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't read config file '%s': %w", filePath, err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	cfg := &ocmConfig.Config{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't parse config file '%s': %w", filePath, err)
	}

	return cfg, nil
}
