package ocm

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

// GetSupportRoleArnForCluster retrieves the AWS support role ARN for a cluster
func GetSupportRoleArnForCluster(ocmClient *sdk.Connection, clusterID string) (string, error) {
	clusterResponse, err := ocmClient.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
	if err != nil {
		return "", err
	}

	// If the cluster is Hypershift, get the ARN from the cluster response body
	if clusterResponse.Body().Hypershift().Enabled() {
		return clusterResponse.Body().AWS().STS().SupportRoleARN(), nil
	}

	// For non-hypershift, the ARN is in the accountclaim
	liveResponse, err := ocmClient.ClustersMgmt().V1().Clusters().Cluster(clusterID).Resources().Live().Get().Send()
	if err != nil {
		return "", err
	}

	respBody := liveResponse.Body().Resources()
	if awsAccountClaim, ok := respBody["aws_account_claim"]; ok {
		var claimJson map[string]interface{}
		err := json.Unmarshal([]byte(awsAccountClaim), &claimJson)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal account claim JSON: %w", err)
		}

		if spec, ok := claimJson["spec"]; ok {
			if supportRoleArn, ok := spec.(map[string]interface{})["supportRoleARN"]; ok {
				return supportRoleArn.(string), nil
			}
		}

		return "", fmt.Errorf("unable to get role arn from claim JSON")
	}

	return "", fmt.Errorf("cluster does not have AccountClaim")
}

// GetAWSAccountIdForCluster extracts the AWS account ID from a cluster's support role ARN
func GetAWSAccountIdForCluster(ocmClient *sdk.Connection, clusterID string) (string, error) {
	roleArn, err := GetSupportRoleArnForCluster(ocmClient, clusterID)
	if err != nil {
		return "", err
	}

	awsRoleArn, err := arn.Parse(roleArn)
	if err != nil {
		return "", err
	}

	return awsRoleArn.AccountID, nil
}
