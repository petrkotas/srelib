package ocm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// GenerateQuery returns an OCM search query to retrieve all clusters matching an expression
func GenerateQuery(clusterIdentifier string) string {
	// Based on the format of the clusterIdentifier, we can know what it is
	if regexp.MustCompile(`^[0-9a-z]{32}$`).MatchString(clusterIdentifier) {
		return strings.TrimSpace(fmt.Sprintf("(id = '%[1]s')", clusterIdentifier))
	} else if _, err := uuid.Parse(clusterIdentifier); err == nil {
		return strings.TrimSpace(fmt.Sprintf("(external_id = '%[1]s')", clusterIdentifier))
	} else {
		return strings.TrimSpace(fmt.Sprintf("(display_name like '%[1]s')", clusterIdentifier))
	}
}

// ApplyFilters retrieves clusters in OCM which match the filters given
func ApplyFilters(ocmClient *sdk.Connection, filters []string) ([]*cmv1.Cluster, error) {
	if len(filters) < 1 {
		return nil, nil
	}

	for k, v := range filters {
		filters[k] = fmt.Sprintf("(%s)", v)
	}

	requestSize := 50
	fullFilters := strings.Join(filters, " and ")

	request := ocmClient.ClustersMgmt().V1().Clusters().List().Search(fullFilters).Size(requestSize)
	response, err := request.Send()
	if err != nil {
		return nil, err
	}

	items := response.Items().Slice()
	for response.Size() >= requestSize {
		request.Page(response.Page() + 1)
		response, err = request.Send()
		if err != nil {
			return nil, err
		}
		items = append(items, response.Items().Slice()...)
	}

	return items, nil
}
