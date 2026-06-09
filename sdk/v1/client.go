package v1

import (
	"errors"
	"net/rpc"

	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// =================================================
// Client implementation
// =================================================

type RPCClient struct {
	Client *rpc.Client
}

// =================================================
// OCM Cluster Operations
// =================================================

func (c *RPCClient) GetCluster(key string) (*cmv1.Cluster, error) {
	args := &GetClusterArgs{Key: key}
	reply := &GetClusterReply{}

	err := c.Client.Call("Plugin.GetCluster", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return DeserializeCluster(reply.ClusterJSON)
}

func (c *RPCClient) GetClusterAnyStatus(clusterId string) (*cmv1.Cluster, error) {
	args := &GetClusterAnyStatusArgs{ClusterID: clusterId}
	reply := &GetClusterAnyStatusReply{}

	err := c.Client.Call("Plugin.GetClusterAnyStatus", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return DeserializeCluster(reply.ClusterJSON)
}

func (c *RPCClient) GetClusters(clusterIds []string) ([]*cmv1.Cluster, error) {
	args := &GetClustersArgs{ClusterIDs: clusterIds}
	reply := &GetClustersReply{}

	err := c.Client.Call("Plugin.GetClusters", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return DeserializeClusters(reply.ClustersJSON)
}

func (c *RPCClient) IsClusterCCS(cluster *cmv1.Cluster) (bool, error) {
	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		return false, err
	}

	args := &IsClusterCCSArgs{ClusterJSON: clusterJSON}
	reply := &IsClusterCCSReply{}

	err = c.Client.Call("Plugin.IsClusterCCS", args, reply)
	if err != nil {
		return false, err
	}

	if reply.Error != "" {
		return false, errors.New(reply.Error)
	}

	return reply.IsCCS, nil
}

func (c *RPCClient) IsHostedCluster(cluster *cmv1.Cluster) (bool, error) {
	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		return false, err
	}

	args := &IsHostedClusterArgs{ClusterJSON: clusterJSON}
	reply := &IsHostedClusterReply{}

	err = c.Client.Call("Plugin.IsHostedCluster", args, reply)
	if err != nil {
		return false, err
	}

	if reply.Error != "" {
		return false, errors.New(reply.Error)
	}

	return reply.IsHosted, nil
}

func (c *RPCClient) GetManagementCluster(cluster *cmv1.Cluster) (*cmv1.Cluster, error) {
	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		return nil, err
	}

	args := &GetManagementClusterArgs{ClusterJSON: clusterJSON}
	reply := &GetManagementClusterReply{}

	err = c.Client.Call("Plugin.GetManagementCluster", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return DeserializeCluster(reply.ManagementClusterJSON)
}

// =================================================
// OCM Subscription & Organization
// =================================================

func (c *RPCClient) GetSubscription(key string) (*amsv1.Subscription, error) {
	args := &GetSubscriptionArgs{Key: key}
	reply := &GetSubscriptionReply{}

	err := c.Client.Call("Plugin.GetSubscription", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return DeserializeSubscription(reply.SubscriptionJSON)
}

func (c *RPCClient) GetOrganization(orgId string) (*amsv1.Organization, error) {
	args := &GetOrganizationArgs{OrgID: orgId}
	reply := &GetOrganizationReply{}

	err := c.Client.Call("Plugin.GetOrganization", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return DeserializeOrganization(reply.OrganizationJSON)
}

func (c *RPCClient) GetOrgFromClusterID(clusterId string) (string, error) {
	args := &GetOrgFromClusterIDArgs{ClusterID: clusterId}
	reply := &GetOrgFromClusterIDReply{}

	err := c.Client.Call("Plugin.GetOrgFromClusterID", args, reply)
	if err != nil {
		return "", err
	}

	if reply.Error != "" {
		return "", errors.New(reply.Error)
	}

	return reply.OrgID, nil
}

// =================================================
// AWS Account Operations
// =================================================

func (c *RPCClient) GetSupportRoleArnForCluster(cluster *cmv1.Cluster) (string, error) {
	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		return "", err
	}

	args := &GetSupportRoleArnForClusterArgs{ClusterJSON: clusterJSON}
	reply := &GetSupportRoleArnForClusterReply{}

	err = c.Client.Call("Plugin.GetSupportRoleArnForCluster", args, reply)
	if err != nil {
		return "", err
	}

	if reply.Error != "" {
		return "", errors.New(reply.Error)
	}

	return reply.Arn, nil
}

func (c *RPCClient) GetAWSAccountIdForCluster(cluster *cmv1.Cluster) (string, error) {
	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		return "", err
	}

	args := &GetAWSAccountIdForClusterArgs{ClusterJSON: clusterJSON}
	reply := &GetAWSAccountIdForClusterReply{}

	err = c.Client.Call("Plugin.GetAWSAccountIdForCluster", args, reply)
	if err != nil {
		return "", err
	}

	if reply.Error != "" {
		return "", errors.New(reply.Error)
	}

	return reply.AccountID, nil
}
