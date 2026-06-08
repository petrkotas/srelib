package i1

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/petrkotas/srelib/internal/ocm"
)

// Client is the V1 implementation of the v1 Client interface.
type Client struct {
	// common logger used internally by the library
	Logger hclog.Logger
	// OCM connection
	ocmConn *sdk.Connection
}

// OCM Connection Management

func (c *Client) CreateOCMConnection(url string) error {
	var conn *sdk.Connection
	var err error

	if url == "" {
		conn, err = ocm.CreateConnection()
	} else {
		conn, err = ocm.CreateConnectionWithUrl(url)
	}

	if err != nil {
		return err
	}

	c.ocmConn = conn
	return nil
}

func (c *Client) GetOCMConnection() (*sdk.Connection, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return c.ocmConn, nil
}

func (c *Client) CloseOCMConnection() error {
	if c.ocmConn != nil {
		err := c.ocmConn.Close()
		c.ocmConn = nil
		return err
	}
	return nil
}

// OCM Cluster Operations

func (c *Client) GetCluster(key string) (*cmv1.Cluster, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetCluster(c.ocmConn, key)
}

func (c *Client) GetClusterAnyStatus(clusterId string) (*cmv1.Cluster, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetClusterAnyStatus(c.ocmConn, clusterId)
}

func (c *Client) GetClusters(clusterIds []string) ([]*cmv1.Cluster, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetClusters(c.ocmConn, clusterIds), nil
}

func (c *Client) IsClusterCCS(cluster *cmv1.Cluster) (bool, error) {
	if c.ocmConn == nil {
		return false, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.IsClusterCCS(c.ocmConn, cluster.ID())
}

func (c *Client) IsHostedCluster(cluster *cmv1.Cluster) (bool, error) {
	if c.ocmConn == nil {
		return false, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.IsHostedCluster(cluster.ID(), c.ocmConn)
}

func (c *Client) GetManagementCluster(cluster *cmv1.Cluster) (*cmv1.Cluster, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetManagementCluster(cluster.ID(), c.ocmConn)
}

// OCM Subscription & Organization

func (c *Client) GetSubscription(key string) (*amsv1.Subscription, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetSubscription(c.ocmConn, key)
}

func (c *Client) GetOrganization(orgId string) (*amsv1.Organization, error) {
	if c.ocmConn == nil {
		return nil, fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetOrganization(c.ocmConn, orgId)
}

func (c *Client) GetOrgFromClusterID(clusterId string) (string, error) {
	if c.ocmConn == nil {
		return "", fmt.Errorf("OCM connection not initialized")
	}
	cluster, err := ocm.GetCluster(c.ocmConn, clusterId)
	if err != nil {
		return "", err
	}
	return ocm.GetOrgFromClusterID(c.ocmConn, cluster)
}

// OCM AWS Integration

func (c *Client) GetSupportRoleArnForCluster(cluster *cmv1.Cluster) (string, error) {
	if c.ocmConn == nil {
		return "", fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetSupportRoleArnForCluster(c.ocmConn, cluster.ID())
}

func (c *Client) GetAWSAccountIdForCluster(cluster *cmv1.Cluster) (string, error) {
	if c.ocmConn == nil {
		return "", fmt.Errorf("OCM connection not initialized")
	}
	return ocm.GetAWSAccountIdForCluster(c.ocmConn, cluster.ID())
}

// OCM Configuration

func (c *Client) LoadOCMConfig(filePath string) error {
	var err error
	if filePath == "" {
		_, err = ocm.LoadOCMConfig()
	} else {
		_, err = ocm.LoadOCMConfigFromPath(filePath)
	}
	return err
}

func (c *Client) GetOCMConfigLocation() (string, error) {
	return ocm.GetOCMConfigLocation()
}
