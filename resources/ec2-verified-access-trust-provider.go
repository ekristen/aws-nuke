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

		for _, trustProvider := range resp.VerifiedAccessTrustProviders {
			resources = append(resources, &EC2VerifiedAccessTrustProvider{
				svc:             svc,
				trustProvider:   &trustProvider,
				ID:              trustProvider.VerifiedAccessTrustProviderId,
				Type:            ptr.String(string(trustProvider.TrustProviderType)),
				Description:     trustProvider.Description,
				CreationTime:    trustProvider.CreationTime,
				LastUpdatedTime: trustProvider.LastUpdatedTime,
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
	trustProvider   *ec2types.VerifiedAccessTrustProvider
	ID              *string
	Type            *string
	Description     *string
	CreationTime    *string
	LastUpdatedTime *string
}

func (r *EC2VerifiedAccessTrustProvider) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: r.trustProvider.VerifiedAccessTrustProviderId,
	}

	_, err := r.svc.DeleteVerifiedAccessTrustProvider(ctx, params)
	return err
}

func (r *EC2VerifiedAccessTrustProvider) Properties() types.Properties {
	properties := types.NewPropertiesFromStruct(r)
	
	for _, tag := range r.trustProvider.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	
	return properties
}

func (r *EC2VerifiedAccessTrustProvider) String() string {
	return *r.trustProvider.VerifiedAccessTrustProviderId
}
