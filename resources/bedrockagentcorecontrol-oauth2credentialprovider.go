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

const BedrockAgentCoreOauth2CredentialProviderResource = "BedrockAgentCoreOauth2CredentialProvider"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreOauth2CredentialProviderResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreOauth2CredentialProvider{},
		Lister:   &BedrockAgentCoreOauth2CredentialProviderLister{},
	})
}

type BedrockAgentCoreOauth2CredentialProviderLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreOauth2CredentialProviderLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListOauth2CredentialProvidersInput{
		MaxResults: aws.Int32(20),
	}

	paginator := bedrockagentcorecontrol.NewListOauth2CredentialProvidersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, provider := range resp.CredentialProviders {
			// Get tags for the OAuth2 credential provider
			var tags map[string]string
			if provider.CredentialProviderArn != nil {
				tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
					ResourceArn: provider.CredentialProviderArn,
				})
				if err != nil {
					opts.Logger.Warnf("unable to fetch tags for OAuth2 credential provider: %s", *provider.CredentialProviderArn)
				} else {
					tags = tagsResp.Tags
				}
			}

			resources = append(resources, &BedrockAgentCoreOauth2CredentialProvider{
				svc:             svc,
				Name:            provider.Name,
				Vendor:          string(provider.CredentialProviderVendor),
				CreatedTime:     provider.CreatedTime,
				LastUpdatedTime: provider.LastUpdatedTime,
				Tags:            tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreOauth2CredentialProvider struct {
	svc             *bedrockagentcorecontrol.Client
	Name            *string
	Vendor          string
	CreatedTime     *time.Time
	LastUpdatedTime *time.Time
	Tags            map[string]string
}

func (r *BedrockAgentCoreOauth2CredentialProvider) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteOauth2CredentialProvider(ctx, &bedrockagentcorecontrol.DeleteOauth2CredentialProviderInput{
		Name: r.Name,
	})

	return err
}

func (r *BedrockAgentCoreOauth2CredentialProvider) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreOauth2CredentialProvider) String() string {
	return *r.Name
}
