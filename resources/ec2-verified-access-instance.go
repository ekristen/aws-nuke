package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VerifiedAccessInstanceResource = "EC2VerifiedAccessInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VerifiedAccessInstanceResource,
		Scope:    nuke.Account,
		Resource: &EC2VerifiedAccessInstance{},
		Lister:   &EC2VerifiedAccessInstanceLister{},
		DependsOn: []string{
			EC2VerifiedAccessGroupResource,
			EC2VerifiedAccessEndpointResource,
		},
	})
}

type EC2VerifiedAccessInstanceLister struct{}

func (l *EC2VerifiedAccessInstanceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	params := &ec2.DescribeVerifiedAccessInstancesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVerifiedAccessInstances(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.VerifiedAccessInstances {
			instance := &resp.VerifiedAccessInstances[i]
			trustProviders := make([]string, 0)
			if instance.VerifiedAccessTrustProviders != nil {
				for _, tp := range instance.VerifiedAccessTrustProviders {
					if tp.VerifiedAccessTrustProviderId != nil {
						trustProviders = append(trustProviders, *tp.VerifiedAccessTrustProviderId)
					}
				}
			}

			resources = append(resources, &EC2VerifiedAccessInstance{
				svc:             svc,
				ID:              instance.VerifiedAccessInstanceId,
				Description:     instance.Description,
				CreationTime:    instance.CreationTime,
				LastUpdatedTime: instance.LastUpdatedTime,
				TrustProviders:  &trustProviders,
				Tags:            instance.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2VerifiedAccessInstance struct {
	svc             *ec2.Client
	ID              *string           `description:"The unique identifier of the Verified Access instance"`
	Description     *string           `description:"A description for the Verified Access instance"`
	CreationTime    *string           `description:"The timestamp when the Verified Access instance was created"`
	LastUpdatedTime *string           `description:"The timestamp when the Verified Access instance was last updated"`
	TrustProviders  *[]string         `description:"A list of trust provider IDs attached to the Verified Access instance"`
	Tags            []ec2types.Tag    `description:"The tags associated with the Verified Access instance"`
}

func (r *EC2VerifiedAccessInstance) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessInstanceInput{
		VerifiedAccessInstanceId: r.ID,
	}

	_, err := r.svc.DeleteVerifiedAccessInstance(ctx, params)
	return err
}

func (r *EC2VerifiedAccessInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2VerifiedAccessInstance) String() string {
	return *r.ID
}
