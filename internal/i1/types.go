package i1

import (
	"errors"
	"sync"

	"github.com/hashicorp/go-hclog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/petrkotas/srelib/internal/aws"
	"github.com/petrkotas/srelib/internal/ocm"
)

// Client is the V1 implementation of the v1 Client interface.
type Client struct {
	// common logger used internally by the library
	Logger hclog.Logger
	// OCM connection
	ocmConn *sdk.Connection
	// once ensures thread-safe lazy initialization
	once sync.Once
	// connErr caches connection creation errors
	connErr error
}

func NewClient(logger hclog.Logger) (*Client, error) {
	var conn *sdk.Connection
	var err error

	conn, err = ocm.CreateConnection()
	if err != nil {
		return nil, err
	}

	return &Client{
		Logger:  logger,
		ocmConn: conn,
	}, nil
}

// NewTestClient creates a client specifically for testing purposes.
// This bypasses the OCM config file requirement and uses environment variables:
// - OCM_URL: URL of the OCM API (or mock server)
// - OCM_TOKEN: Authentication token
//
// This function should ONLY be used in test code.
func NewTestClient(logger hclog.Logger) (*Client, error) {
	var conn *sdk.Connection
	var err error

	conn, err = ocm.CreateTestConnection()
	if err != nil {
		return nil, err
	}

	return &Client{
		Logger:  logger,
		ocmConn: conn,
	}, nil
}

// Close closes the OCM connection and cleans up resources.
func (c *Client) Close() error {
	if c.ocmConn != nil {
		err := c.ocmConn.Close()
		c.ocmConn = nil
		return err
	}
	return nil
}

// OCM Cluster Operations

func (c *Client) GetCluster(key string) (*cmv1.Cluster, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}

	return ocm.GetCluster(c.ocmConn, key)
}

func (c *Client) GetClusterAnyStatus(clusterId string) (*cmv1.Cluster, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}
	return ocm.GetClusterAnyStatus(c.ocmConn, clusterId)
}

func (c *Client) GetClusters(clusterIds []string) ([]*cmv1.Cluster, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}
	return ocm.GetClusters(c.ocmConn, clusterIds), nil
}

func (c *Client) GetManagementCluster(clusterId string) (*cmv1.Cluster, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}
	return ocm.GetManagementCluster(clusterId, c.ocmConn)
}

// OCM Subscription & Organization

func (c *Client) GetSubscription(key string) (*amsv1.Subscription, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}
	return ocm.GetSubscription(c.ocmConn, key)
}

func (c *Client) GetOrganization(orgId string) (*amsv1.Organization, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}
	return ocm.GetOrganization(c.ocmConn, orgId)
}

// AWS Account Operations

func (c *Client) GetSupportRoleArnForCluster(clusterId string) (string, error) {
	if c.connErr != nil {
		return "", errors.New("OCM connection error")
	}
	return aws.GetSupportRoleArnFromCluster(c.ocmConn, clusterId)
}

func (c *Client) GetAWSAccountIdForCluster(clusterId string) (string, error) {
	if c.connErr != nil {
		return "", errors.New("OCM connection error")
	}
	return aws.GetAccountIdFromCluster(c.ocmConn, clusterId)
}
