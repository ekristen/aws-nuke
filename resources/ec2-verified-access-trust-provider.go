package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VerifiedAccessTrustProviderResource = "EC2VerifiedAccessTrustProvider"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VerifiedAccessTrustProviderResource,
		Scope:    nuke.Account,
		Resource: &EC2VerifiedAccessTrustProvider{},
		Lister:   &EC2VerifiedAccessTrustProviderLister{},
	})
}

type EC2VerifiedAccessTrustProviderLister struct{}

func (l *EC2VerifiedAccessTrustProviderLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	params := &ec2.DescribeVerifiedAccessTrustProvidersInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVerifiedAccessTrustProviders(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.VerifiedAccessTrustProviders {
			trustProvider := &resp.VerifiedAccessTrustProviders[i]
			resources = append(resources, &EC2VerifiedAccessTrustProvider{
				svc:             svc,
				ID:              trustProvider.VerifiedAccessTrustProviderId,
				Type:            ptr.String(string(trustProvider.TrustProviderType)),
				Description:     trustProvider.Description,
				CreationTime:    trustProvider.CreationTime,
				LastUpdatedTime: trustProvider.LastUpdatedTime,
				Tags:            trustProvider.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2VerifiedAccessTrustProvider struct {
	svc             *ec2.Client
	ID              *string        `description:"The unique identifier of the Verified Access trust provider"`
	Type            *string        `description:"The type of trust provider (user, device, or oidc)"`
	Description     *string        `description:"A description for the Verified Access trust provider"`
	CreationTime    *string        `description:"The timestamp when the Verified Access trust provider was created"`
	LastUpdatedTime *string        `description:"The timestamp when the Verified Access trust provider was last updated"`
	Tags            []ec2types.Tag `description:"The tags associated with the Verified Access trust provider"`
}

func (r *EC2VerifiedAccessTrustProvider) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: r.ID,
	}

	_, err := r.svc.DeleteVerifiedAccessTrustProvider(ctx, params)
	return err
}

func (r *EC2VerifiedAccessTrustProvider) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2VerifiedAccessTrustProvider) String() string {
	return *r.ID
}
