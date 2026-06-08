package ocm

import (
	"fmt"
	"log"
	"strings"

	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const ClusterServiceClusterSearch = "id = '%s' or name = '%s' or external_id = '%s'"

// GetCluster allows getting a single cluster with any identifier (displayname, ID, or external ID)
func GetCluster(connection *sdk.Connection, key string) (cluster *cmv1.Cluster, err error) {
	subsResource := connection.AccountsMgmt().V1().Subscriptions()
	clustersResource := connection.ClustersMgmt().V1().Clusters()

	subsSearch := fmt.Sprintf(
		"(display_name = '%s' or cluster_id = '%s' or external_cluster_id = '%s') and "+
			"status in ('Reserved', 'Active')",
		key, key, key,
	)
	subsListResponse, err := subsResource.List().
		Search(subsSearch).
		Size(1).
		Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve subscription for key '%s': %w", key, err)
	}

	subsTotal := subsListResponse.Total()
	if subsTotal == 1 {
		id, ok := subsListResponse.Items().Slice()[0].GetClusterID()
		if ok {
			clusterGetResponse, err := clustersResource.Cluster(id).Get().Send()
			if err != nil {
				return nil, fmt.Errorf("can't retrieve cluster for key '%s': %w", key, err)
			}
			return clusterGetResponse.Body(), nil
		}
	}

	if subsTotal > 1 {
		return nil, fmt.Errorf("there are %d subscriptions with cluster identifier or name '%s'", subsTotal, key)
	}

	clustersSearch := fmt.Sprintf("id = '%s' or name = '%s' or external_id = '%s'", key, key, key)
	clustersListResponse, err := clustersResource.List().
		Search(clustersSearch).
		Size(1).
		Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve clusters for key '%s': %w", key, err)
	}

	clustersTotal := clustersListResponse.Total()
	if clustersTotal == 1 {
		return clustersListResponse.Items().Slice()[0], nil
	}

	if clustersTotal > 1 {
		return nil, fmt.Errorf("there are %d clusters with identifier or name '%s'", clustersTotal, key)
	}

	return nil, fmt.Errorf("there are no subscriptions or clusters with identifier or name '%s'", key)
}

// GetClusterAnyStatus returns an OCM cluster object given an OCM connection and cluster id
// (internal id, external id, and name all supported).
func GetClusterAnyStatus(conn *sdk.Connection, clusterId string) (*cmv1.Cluster, error) {
	clustersSearch := fmt.Sprintf(ClusterServiceClusterSearch, clusterId, clusterId, clusterId)
	clustersListResponse, err := conn.ClustersMgmt().V1().Clusters().List().Search(clustersSearch).Size(1).Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve clusters for clusterId '%s': %w", clusterId, err)
	}

	clustersTotal := clustersListResponse.Total()
	if clustersTotal == 1 {
		return clustersListResponse.Items().Slice()[0], nil
	}

	return nil, fmt.Errorf("there are %d clusters with identifier or name '%s', expected 1", clustersTotal, clusterId)
}

// GetClusters retrieves multiple clusters by their IDs
func GetClusters(ocmClient *sdk.Connection, clusterIds []string) []*cmv1.Cluster {
	for i, id := range clusterIds {
		clusterIds[i] = GenerateQuery(id)
	}

	clusters, err := ApplyFilters(ocmClient, []string{strings.Join(clusterIds, " or ")})
	if err != nil {
		log.Fatalf("error while retrieving cluster(s) from ocm: %s", err)
	}

	return clusters
}

// GetInternalClusterID converts any cluster identifier to the internal cluster ID
func GetInternalClusterID(ocmClient *sdk.Connection, clusterIdentifier string) (string, error) {
	cluster, err := GetCluster(ocmClient, clusterIdentifier)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster: %w", err)
	}

	return cluster.ID(), nil
}

// IsClusterCCS checks if a cluster is a Customer Cloud Subscription cluster
func IsClusterCCS(ocmClient *sdk.Connection, clusterID string) (bool, error) {
	clusterResponse, err := ocmClient.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
	if err != nil {
		return false, err
	}

	cluster := clusterResponse.Body()
	return cluster.CCS().Enabled(), nil
}

// IsHostedCluster checks if a cluster is a Hypershift/HCP hosted cluster
func IsHostedCluster(clusterID string, conn *sdk.Connection) (bool, error) {
	cluster := conn.ClustersMgmt().V1().Clusters().Cluster(clusterID)
	res, err := cluster.Get().Send()
	if err != nil {
		return false, err
	}

	return res.Body().Hypershift().Enabled(), nil
}

// GetManagementCluster returns the OCM Cluster object for the management cluster of a hosted cluster
func GetManagementCluster(clusterId string, conn *sdk.Connection) (*cmv1.Cluster, error) {
	hypershiftResp, err := conn.ClustersMgmt().V1().Clusters().
		Cluster(clusterId).
		Hypershift().
		Get().
		Send()
	if err != nil {
		return nil, err
	}

	if mgmtClusterName, ok := hypershiftResp.Body().GetManagementCluster(); ok {
		return GetClusterAnyStatus(conn, mgmtClusterName)
	}

	return nil, fmt.Errorf("no management cluster found for %s", clusterId)
}

// GetServiceCluster returns the hypershift Service Cluster object for a provided HCP clusterId
func GetServiceCluster(clusterId string, conn *sdk.Connection) (*cmv1.Cluster, error) {
	var svcClusterName, mgmtClusterName string

	hypershiftResp, err := conn.ClustersMgmt().V1().Clusters().
		Cluster(clusterId).
		Hypershift().
		Get().
		Send()
	if err != nil {
		return nil, err
	}

	if hypershiftResp != nil {
		mgmtClusterName = hypershiftResp.Body().ManagementCluster()
	}

	if mgmtClusterName == "" {
		return nil, fmt.Errorf("failed to lookup management cluster for cluster %s", clusterId)
	}

	ofmResp, err := conn.OSDFleetMgmt().V1().ManagementClusters().
		List().
		Parameter("search", fmt.Sprintf("name='%s'", mgmtClusterName)).
		Send()
	if err != nil {
		return nil, fmt.Errorf("failed to get the fleet manager information for management cluster %s", mgmtClusterName)
	}

	if kind := ofmResp.Items().Get(0).Parent().Kind(); kind == "ServiceCluster" {
		svcClusterName = ofmResp.Items().Get(0).Parent().Name()
	}

	svcCluster, err := GetClusterAnyStatus(conn, svcClusterName)
	if err != nil {
		return nil, err
	}

	return svcCluster, nil
}

// IsManagementCluster checks if a cluster is a management cluster
func IsManagementCluster(clusterID string, conn *sdk.Connection) (bool, error) {
	collection := conn.ClustersMgmt().V1().Clusters()
	list, err := collection.List().
		Parameter("search", fmt.Sprintf("hypershift.management_cluster='%s'", clusterID)).
		Size(1).
		Send()
	if err != nil {
		return false, fmt.Errorf("failed to check if cluster %s is a management cluster: %w", clusterID, err)
	}

	return list.Total() >= 1, nil
}

// GetClusterLimitedSupportReasons retrieves limited support reasons for a cluster
func GetClusterLimitedSupportReasons(connection *sdk.Connection, clusterID string) ([]*cmv1.LimitedSupportReason, error) {
	limitedSupportReasons, err := connection.ClustersMgmt().V1().
		Clusters().
		Cluster(clusterID).
		LimitedSupportReasons().
		List().
		Send()
	if err != nil {
		return nil, fmt.Errorf("failed to get limited support reasons: %s", err)
	}

	return limitedSupportReasons.Items().Slice(), nil
}

// GetHCPNamespace gets the HCP namespace for a hosted cluster
func GetHCPNamespace(clusterId string, conn *sdk.Connection) (string, error) {
	hypershiftResp, err := conn.ClustersMgmt().V1().Clusters().
		Cluster(clusterId).
		Hypershift().
		Get().
		Send()
	if err != nil {
		return "", err
	}

	if namespace, ok := hypershiftResp.Body().GetHCPNamespace(); ok {
		return namespace, nil
	}

	return "", fmt.Errorf("no hcp namespace found for %s", clusterId)
}
