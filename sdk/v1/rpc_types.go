package v1

import (
	"encoding/json"

	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// =================================================
// RPC Argument and Reply Types
// These types are used for RPC communication and
// handle serialization of OCM SDK types via JSON
// =================================================

// GetClusterArgs are the arguments for GetCluster RPC
type GetClusterArgs struct {
	Key string
}

// GetClusterReply is the reply for GetCluster RPC
type GetClusterReply struct {
	ClusterJSON []byte
	Error       string
}

// GetClusterAnyStatusArgs are the arguments for GetClusterAnyStatus RPC
type GetClusterAnyStatusArgs struct {
	ClusterID string
}

// GetClusterAnyStatusReply is the reply for GetClusterAnyStatus RPC
type GetClusterAnyStatusReply struct {
	ClusterJSON []byte
	Error       string
}

// GetClustersArgs are the arguments for GetClusters RPC
type GetClustersArgs struct {
	ClusterIDs []string
}

// GetClustersReply is the reply for GetClusters RPC
type GetClustersReply struct {
	ClustersJSON [][]byte
	Error        string
}

// IsClusterCCSArgs are the arguments for IsClusterCCS RPC
type IsClusterCCSArgs struct {
	ClusterJSON []byte
}

// IsClusterCCSReply is the reply for IsClusterCCS RPC
type IsClusterCCSReply struct {
	IsCCS bool
	Error string
}

// IsHostedClusterArgs are the arguments for IsHostedCluster RPC
type IsHostedClusterArgs struct {
	ClusterJSON []byte
}

// IsHostedClusterReply is the reply for IsHostedCluster RPC
type IsHostedClusterReply struct {
	IsHosted bool
	Error    string
}

// GetManagementClusterArgs are the arguments for GetManagementCluster RPC
type GetManagementClusterArgs struct {
	ClusterID string
}

// GetManagementClusterReply is the reply for GetManagementCluster RPC
type GetManagementClusterReply struct {
	ManagementClusterJSON []byte
	Error                 string
}

// GetSubscriptionArgs are the arguments for GetSubscription RPC
type GetSubscriptionArgs struct {
	Key string
}

// GetSubscriptionReply is the reply for GetSubscription RPC
type GetSubscriptionReply struct {
	SubscriptionJSON []byte
	Error            string
}

// GetOrganizationArgs are the arguments for GetOrganization RPC
type GetOrganizationArgs struct {
	OrgID string
}

// GetOrganizationReply is the reply for GetOrganization RPC
type GetOrganizationReply struct {
	OrganizationJSON []byte
	Error            string
}

// GetOrgFromClusterIDArgs are the arguments for GetOrgFromClusterID RPC
type GetOrgFromClusterIDArgs struct {
	ClusterID string
}

// GetOrgFromClusterIDReply is the reply for GetOrgFromClusterID RPC
type GetOrgFromClusterIDReply struct {
	OrgID string
	Error string
}

// GetSupportRoleArnForClusterArgs are the arguments for GetSupportRoleArnForCluster RPC
type GetSupportRoleArnForClusterArgs struct {
	ClusterID string
}

// GetSupportRoleArnForClusterReply is the reply for GetSupportRoleArnForCluster RPC
type GetSupportRoleArnForClusterReply struct {
	Arn   string
	Error string
}

// GetAWSAccountIdForClusterArgs are the arguments for GetAWSAccountIdForCluster RPC
type GetAWSAccountIdForClusterArgs struct {
	ClusterID string
}

// GetAWSAccountIdForClusterReply is the reply for GetAWSAccountIdForCluster RPC
type GetAWSAccountIdForClusterReply struct {
	AccountID string
	Error     string
}

// =================================================
// Helper functions for serialization
// =================================================

// SerializeCluster converts a Cluster to JSON bytes
func SerializeCluster(cluster *cmv1.Cluster) ([]byte, error) {
	if cluster == nil {
		return nil, nil
	}
	return json.Marshal(cluster)
}

// DeserializeCluster converts JSON bytes to a Cluster
func DeserializeCluster(data []byte) (*cmv1.Cluster, error) {
	if data == nil {
		return nil, nil
	}
	cluster := &cmv1.Cluster{}
	err := json.Unmarshal(data, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// SerializeClusters converts a slice of Clusters to JSON bytes
func SerializeClusters(clusters []*cmv1.Cluster) ([][]byte, error) {
	result := make([][]byte, len(clusters))
	for i, cluster := range clusters {
		data, err := SerializeCluster(cluster)
		if err != nil {
			return nil, err
		}
		result[i] = data
	}
	return result, nil
}

// DeserializeClusters converts JSON bytes to a slice of Clusters
func DeserializeClusters(data [][]byte) ([]*cmv1.Cluster, error) {
	result := make([]*cmv1.Cluster, len(data))
	for i, clusterData := range data {
		cluster, err := DeserializeCluster(clusterData)
		if err != nil {
			return nil, err
		}
		result[i] = cluster
	}
	return result, nil
}

// SerializeSubscription converts a Subscription to JSON bytes
func SerializeSubscription(sub *amsv1.Subscription) ([]byte, error) {
	if sub == nil {
		return nil, nil
	}
	return json.Marshal(sub)
}

// DeserializeSubscription converts JSON bytes to a Subscription
func DeserializeSubscription(data []byte) (*amsv1.Subscription, error) {
	if data == nil {
		return nil, nil
	}
	sub := &amsv1.Subscription{}
	err := json.Unmarshal(data, sub)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// SerializeOrganization converts an Organization to JSON bytes
func SerializeOrganization(org *amsv1.Organization) ([]byte, error) {
	if org == nil {
		return nil, nil
	}
	return json.Marshal(org)
}

// DeserializeOrganization converts JSON bytes to an Organization
func DeserializeOrganization(data []byte) (*amsv1.Organization, error) {
	if data == nil {
		return nil, nil
	}
	org := &amsv1.Organization{}
	err := json.Unmarshal(data, org)
	if err != nil {
		return nil, err
	}
	return org, nil
}
