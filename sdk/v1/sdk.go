package v1

import (
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// Client is the interface with actions the plugin provides.
type Client interface {

	// OCM Cluster Operations

	// GetCluster allows getting a single cluster with any identifier
	// (displayname, ID, or external ID)
	GetCluster(key string) (*cmv1.Cluster, error)
	GetClusterAnyStatus(clusterId string) (*cmv1.Cluster, error)
	GetClusters(clusterIds []string) ([]*cmv1.Cluster, error)
	GetManagementCluster(clusterId string) (*cmv1.Cluster, error)

	// OCM Subscription & Organization
	GetSubscription(key string) (*amsv1.Subscription, error)
	GetOrganization(orgId string) (*amsv1.Organization, error)

	// AWS Account Operations
	GetSupportRoleArnForCluster(clusterId string) (string, error)
	GetAWSAccountIdForCluster(clusterId string) (string, error)
}
