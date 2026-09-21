package resources

import (
	"regexp"

	"github.com/aws/aws-sdk-go/service/quicksight" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/quicksight/quicksightiface"
)

// quickSightDefaultNamespace is the namespace that every QuickSight subscription has.
const quickSightDefaultNamespace = "default"

// quickSightIdentityRegionRe matches the region named by the AccessDeniedException that
// QuickSight returns for an identity operation called outside the identity region:
//
//	Operation is being called from endpoint us-west-2, but your identity region is
//	us-east-1. Please use the us-east-1 endpoint.
var quickSightIdentityRegionRe = regexp.MustCompile(`identity region is ([a-z0-9-]+)`)

// QuickSightIdentityRegion returns the region that owns the account's QuickSight identity,
// or an empty string if it cannot be determined.
//
// A QuickSight subscription is account wide, but users, groups, namespaces and the
// subscription itself can only be modified through the identity region's endpoint. Some
// read operations (DescribeAccountSubscription in particular) answer from any region, so
// without this the same resource is listed once per region and fails to delete in all but
// one of them.
//
// The namespaces carry the region directly. When the call is itself rejected for being
// outside the identity region, AWS names that region in the error message, which answers
// the question just as well.
func QuickSightIdentityRegion(svc quicksightiface.QuickSightAPI, accountID *string) string {
	out, err := svc.ListNamespaces(&quicksight.ListNamespacesInput{
		AwsAccountId: accountID,
	})
	if err != nil {
		if matches := quickSightIdentityRegionRe.FindStringSubmatch(err.Error()); matches != nil {
			return matches[1]
		}
		return ""
	}

	var region string
	for _, namespace := range out.Namespaces {
		if namespace.CapacityRegion == nil || *namespace.CapacityRegion == "" {
			continue
		}
		if namespace.Name != nil && *namespace.Name == quickSightDefaultNamespace {
			return *namespace.CapacityRegion
		}
		if region == "" {
			region = *namespace.CapacityRegion
		}
	}

	return region
}
