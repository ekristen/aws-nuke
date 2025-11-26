package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockAgentCoreAPIKeyCredentialProviderResource = "BedrockAgentCoreAPIKeyCredentialProvider" //nolint:gosec

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreAPIKeyCredentialProviderResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreAPIKeyCredentialProvider{},
		Lister:   &BedrockAgentCoreAPIKeyCredentialProviderLister{},
	})
}

type BedrockAgentCoreAPIKeyCredentialProviderLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreAPIKeyCredentialProviderLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListApiKeyCredentialProvidersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListApiKeyCredentialProvidersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, provider := range resp.CredentialProviders {
			resources = append(resources, &BedrockAgentCoreAPIKeyCredentialProvider{
				svc:                   svc,
				Name:                  provider.Name,
				CredentialProviderArn: provider.CredentialProviderArn,
				CreatedTime:           provider.CreatedTime,
				LastUpdatedTime:       provider.LastUpdatedTime,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreAPIKeyCredentialProvider struct {
	svc                   *bedrockagentcorecontrol.Client
	Name                  *string
	CredentialProviderArn *string
	CreatedTime           *time.Time
	LastUpdatedTime       *time.Time
}

func (r *BedrockAgentCoreAPIKeyCredentialProvider) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteApiKeyCredentialProvider(ctx, &bedrockagentcorecontrol.DeleteApiKeyCredentialProviderInput{
		Name: r.Name,
	})

	return err
}

func (r *BedrockAgentCoreAPIKeyCredentialProvider) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreAPIKeyCredentialProvider) String() string {
	return *r.Name
}
