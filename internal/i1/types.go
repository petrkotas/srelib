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

func (c *Client) IsClusterCCS(cluster *cmv1.Cluster) (bool, error) {
	if c.connErr != nil {
		return false, errors.New("OCM connection error")
	}
	return ocm.IsClusterCCS(c.ocmConn, cluster.ID())
}

func (c *Client) IsHostedCluster(cluster *cmv1.Cluster) (bool, error) {
	if c.connErr != nil {
		return false, errors.New("OCM connection error")
	}
	return ocm.IsHostedCluster(cluster.ID(), c.ocmConn)
}

func (c *Client) GetManagementCluster(cluster *cmv1.Cluster) (*cmv1.Cluster, error) {
	if c.connErr != nil {
		return nil, errors.New("OCM connection error")
	}
	return ocm.GetManagementCluster(cluster.ID(), c.ocmConn)
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

func (c *Client) GetOrgFromClusterID(clusterId string) (string, error) {
	if c.connErr != nil {
		return "", errors.New("OCM connection error")
	}
	cluster, err := ocm.GetCluster(c.ocmConn, clusterId)
	if err != nil {
		return "", err
	}
	return ocm.GetOrgFromClusterID(c.ocmConn, cluster)
}

// AWS Account Operations

func (c *Client) GetSupportRoleArnForCluster(cluster *cmv1.Cluster) (string, error) {
	if c.connErr != nil {
		return "", errors.New("OCM connection error")
	}
	return aws.GetSupportRoleArnFromCluster(c.ocmConn, cluster.ID())
}

func (c *Client) GetAWSAccountIdForCluster(cluster *cmv1.Cluster) (string, error) {
	if c.connErr != nil {
		return "", errors.New("OCM connection error")
	}
	return aws.GetAccountIdFromCluster(c.ocmConn, cluster.ID())
}
