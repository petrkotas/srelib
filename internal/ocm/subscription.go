package ocm

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	amsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
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

// GetOrganization returns an *amsv1.Organization given an subscription ID
func GetOrganization(connection *sdk.Connection, subscriptionId string) (*amsv1.Organization, error) {
	orgResource := connection.AccountsMgmt().V1().Organizations().Organization(subscriptionId)
	orgGetResponse, err := orgResource.Get().Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve organization for key '%s': %w", subscriptionId, err)
	}

	return orgGetResponse.Body(), nil
}
