package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockWorkloadIdentityResource = "BedrockWorkloadIdentity"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockWorkloadIdentityResource,
		Scope:    nuke.Account,
		Resource: &BedrockWorkloadIdentity{},
		Lister:   &BedrockWorkloadIdentityLister{},
	})
}

type BedrockWorkloadIdentityLister struct{}

func (l *BedrockWorkloadIdentityLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &bedrockagentcorecontrol.ListWorkloadIdentitiesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListWorkloadIdentitiesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, identity := range resp.WorkloadIdentities {
			resources = append(resources, &BedrockWorkloadIdentity{
				svc:                 svc,
				Name:                identity.Name,
				WorkloadIdentityArn: identity.WorkloadIdentityArn,
			})
		}
	}

	return resources, nil
}

type BedrockWorkloadIdentity struct {
	svc                 *bedrockagentcorecontrol.Client
	Name                *string
	WorkloadIdentityArn *string
}

func (r *BedrockWorkloadIdentity) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteWorkloadIdentity(ctx, &bedrockagentcorecontrol.DeleteWorkloadIdentityInput{
		Name: r.Name,
	})

	return err
}

func (r *BedrockWorkloadIdentity) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockWorkloadIdentity) String() string {
	return *r.Name
}
