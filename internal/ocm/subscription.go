package ocm

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// GetSubscription allows getting a single subscription with any identifier
// (displayname, ID, internal or external ID)
func GetSubscription(connection *sdk.Connection, key string) (*amsv1.Subscription, error) {
	subsResource := connection.AccountsMgmt().V1().Subscriptions()

	subsSearch := fmt.Sprintf(
		"(display_name = '%s' or cluster_id = '%s' or external_cluster_id = '%s' or id = '%s')",
		key, key, key, key)
	subsListResponse, err := subsResource.List().Parameter("search", subsSearch).Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve subscription for key '%s': %w", key, err)
	}

	subsTotal := subsListResponse.Total()
	if subsTotal == 1 {
		return subsListResponse.Items().Get(0), nil
	}

	if subsTotal > 1 {
		return nil, fmt.Errorf("there are %d subscriptions with cluster identifier or name '%s'", subsTotal, key)
	}

	return nil, fmt.Errorf("there are no subscriptions with identifier or name '%s'", key)
}

// GetOrganization returns an *amsv1.Organization given an OCM cluster name, external id, or internal id as key
func GetOrganization(connection *sdk.Connection, key string) (*amsv1.Organization, error) {
	subscription, err := GetSubscription(connection, key)
	if err != nil {
		return nil, err
	}

	orgResource := connection.AccountsMgmt().V1().Organizations().Organization(subscription.OrganizationID())
	orgGetResponse, err := orgResource.Get().Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve organization for key '%s': %w", key, err)
	}

	return orgGetResponse.Body(), nil
}

// GetOrgFromClusterID returns the organization ID from a cluster object
func GetOrgFromClusterID(ocmClient *sdk.Connection, cluster *cmv1.Cluster) (string, error) {
	sub, err := GetSubFromClusterID(ocmClient, cluster)
	if err != nil {
		return "", err
	}

	return sub.OrganizationID(), nil
}

// GetSubFromClusterID retrieves a subscription from a cluster object
func GetSubFromClusterID(ocmClient *sdk.Connection, cluster *cmv1.Cluster) (*amsv1.Subscription, error) {
	subID, ok := cluster.Subscription().GetID()
	if !ok {
		return nil, fmt.Errorf("failed getting subscription id")
	}

	resp, err := ocmClient.AccountsMgmt().V1().Subscriptions().List().Search(fmt.Sprintf("id like '%s'", subID)).Size(1).Send()
	if err != nil {
		return nil, err
	}

	respSlice := resp.Items().Slice()
	if len(respSlice) > 1 {
		return nil, fmt.Errorf("expected only 1 subscription to be returned")
	} else if len(respSlice) == 0 {
		return nil, fmt.Errorf("subscription not found")
	}

	return respSlice[0], nil
}
