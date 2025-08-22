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

		for _, instance := range resp.VerifiedAccessInstances {
			resources = append(resources, &EC2VerifiedAccessInstance{
				svc:      svc,
				instance: &instance,
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
	svc      *ec2.Client
	instance *ec2types.VerifiedAccessInstance
}

func (r *EC2VerifiedAccessInstance) Remove(ctx context.Context) error {
	params := &ec2.DeleteVerifiedAccessInstanceInput{
		VerifiedAccessInstanceId: r.instance.VerifiedAccessInstanceId,
	}

	_, err := r.svc.DeleteVerifiedAccessInstance(ctx, params)
	return err
}

func (r *EC2VerifiedAccessInstance) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tag := range r.instance.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	properties.Set("ID", r.instance.VerifiedAccessInstanceId)
	properties.Set("Description", r.instance.Description)
	properties.Set("CreationTime", r.instance.CreationTime)
	properties.Set("LastUpdatedTime", r.instance.LastUpdatedTime)

	if r.instance.VerifiedAccessTrustProviders != nil {
		trustProviders := make([]*string, len(r.instance.VerifiedAccessTrustProviders))
		for i, tp := range r.instance.VerifiedAccessTrustProviders {
			trustProviders[i] = tp.VerifiedAccessTrustProviderId
		}
		properties.Set("TrustProviders", trustProviders)
	}

	return properties
}

func (r *EC2VerifiedAccessInstance) String() string {
	return *r.instance.VerifiedAccessInstanceId
}
