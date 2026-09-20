package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SecurityHubV2Resource = "SecurityHubV2"

func init() {
	registry.Register(&registry.Registration{
		Name:     SecurityHubV2Resource,
		Scope:    nuke.Account,
		Resource: &SecurityHubV2{},
		Lister:   &SecurityHubV2Lister{},
	})
}

type SecurityHubV2Lister struct {
	svc SecurityHubV2API
}

// List returns the Security Hub V2 service resource if it is enabled in the current region. Security Hub V2 is a
// per-region opt-in, so an account that has never enabled it (or a region where it is not available) returns nothing
// instead of an error.
func (l *SecurityHubV2Lister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)

	svc := l.svc
	if svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc = securityhub.NewFromConfig(*opts.Config)
	}

	resp, err := svc.DescribeSecurityHubV2(ctx, &securityhub.DescribeSecurityHubV2Input{})
	if err != nil {
		var notFound *securityhubtypes.ResourceNotFoundException
		var invalidAccess *securityhubtypes.InvalidAccessException
		var accessDenied *securityhubtypes.AccessDeniedException
		if errors.As(err, &notFound) || errors.As(err, &invalidAccess) || errors.As(err, &accessDenied) {
			// Security Hub V2 is not enabled or not available in this region
			return resources, nil
		}
		return nil, err
	}

	features := make(map[string]string)
	for name, detail := range resp.Features {
		features[name] = string(detail.FeatureStatus)
	}

	resources = append(resources, &SecurityHubV2{
		svc:          svc,
		HubV2ARN:     resp.HubV2Arn,
		SubscribedAt: resp.SubscribedAt,
		Features:     features,
	})

	return resources, nil
}

// SecurityHubV2 is the account level opt-in for Security Hub V2. Note that disabling it does NOT clean up the service
// managed resources Security Hub V2 provisions on the account's behalf: the _AccessAnalyzerForSecurityHubV2-* IAM
// Access Analyzer survives the disable and still cannot be deleted, so AccessAnalyzer filters it out entirely.
type SecurityHubV2 struct {
	svc          SecurityHubV2API
	HubV2ARN     *string           `description:"The ARN of the Security Hub V2 service resource"`
	SubscribedAt *string           `description:"The date and time the account subscribed to Security Hub V2"`
	Features     map[string]string `property:"tagPrefix=feature" description:"The enablement status of each opt-in feature"`
}

func (r *SecurityHubV2) Remove(ctx context.Context) error {
	_, err := r.svc.DisableSecurityHubV2(ctx, &securityhub.DisableSecurityHubV2Input{})
	return err
}

func (r *SecurityHubV2) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityHubV2) String() string {
	return *r.HubV2ARN
}
