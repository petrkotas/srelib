package v1

import (
	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// Client is the interface with actions the plugin provides.
type Client interface {
	// OCM Connection Management
	CreateOCMConnection(url string) error
	GetOCMConnection() (*sdk.Connection, error)
	CloseOCMConnection() error

	// OCM Cluster Operations
	GetCluster(key string) (*cmv1.Cluster, error)
	GetClusterAnyStatus(clusterId string) (*cmv1.Cluster, error)
	GetClusters(clusterIds []string) ([]*cmv1.Cluster, error)
	IsClusterCCS(cluster *cmv1.Cluster) (bool, error)
	IsHostedCluster(cluster *cmv1.Cluster) (bool, error)
	GetManagementCluster(cluster *cmv1.Cluster) (*cmv1.Cluster, error)

	// OCM Subscription & Organization
	GetSubscription(key string) (*amsv1.Subscription, error)
	GetOrganization(orgId string) (*amsv1.Organization, error)
	GetOrgFromClusterID(clusterId string) (string, error)

	// OCM AWS Integration
	GetSupportRoleArnForCluster(cluster *cmv1.Cluster) (string, error)
	GetAWSAccountIdForCluster(cluster *cmv1.Cluster) (string, error)

	// OCM Configuration
	LoadOCMConfig(filePath string) error
	GetOCMConfigLocation() (string, error)
}
