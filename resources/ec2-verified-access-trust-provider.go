package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"

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

func (l *EC2VerifiedAccessTrustProviderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeVerifiedAccessTrustProvidersInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVerifiedAccessTrustProviders(params)
		if err != nil {
			return nil, err
		}

		for _, trustProvider := range resp.VerifiedAccessTrustProviders {
			resources = append(resources, &EC2VerifiedAccessTrustProvider{
				svc:           svc,
				trustProvider: trustProvider,
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
	svc           *ec2.EC2
	trustProvider *ec2.VerifiedAccessTrustProvider
}

func (r *EC2VerifiedAccessTrustProvider) Remove(_ context.Context) error {
	params := &ec2.DeleteVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: r.trustProvider.VerifiedAccessTrustProviderId,
	}

	_, err := r.svc.DeleteVerifiedAccessTrustProvider(params)
	return err
}

func (r *EC2VerifiedAccessTrustProvider) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tag := range r.trustProvider.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	properties.Set("ID", r.trustProvider.VerifiedAccessTrustProviderId)
	properties.Set("Type", r.trustProvider.TrustProviderType)
	properties.Set("Description", r.trustProvider.Description)
	properties.Set("CreationTime", r.trustProvider.CreationTime)
	properties.Set("LastUpdatedTime", r.trustProvider.LastUpdatedTime)

	return properties
}

func (r *EC2VerifiedAccessTrustProvider) String() string {
	return *r.trustProvider.VerifiedAccessTrustProviderId
}
