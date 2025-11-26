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

const BedrockAPIKeyCredentialProviderResource = "BedrockAPIKeyCredentialProvider" //nolint:gosec

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAPIKeyCredentialProviderResource,
		Scope:    nuke.Account,
		Resource: &BedrockAPIKeyCredentialProvider{},
		Lister:   &BedrockAPIKeyCredentialProviderLister{},
	})
}

type BedrockAPIKeyCredentialProviderLister struct{}

func (l *BedrockAPIKeyCredentialProviderLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

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
			resources = append(resources, &BedrockAPIKeyCredentialProvider{
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

type BedrockAPIKeyCredentialProvider struct {
	svc                   *bedrockagentcorecontrol.Client
	Name                  *string
	CredentialProviderArn *string
	CreatedTime           *time.Time
	LastUpdatedTime       *time.Time
}

func (r *BedrockAPIKeyCredentialProvider) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteApiKeyCredentialProvider(ctx, &bedrockagentcorecontrol.DeleteApiKeyCredentialProviderInput{
		Name: r.Name,
	})

	return err
}

func (r *BedrockAPIKeyCredentialProvider) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAPIKeyCredentialProvider) String() string {
	return *r.Name
}
