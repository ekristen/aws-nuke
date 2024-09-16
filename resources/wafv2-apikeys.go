package resources

import (
	"context"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/wafv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFv2APIKeyResource = "WAFv2APIKey" // #nosec G101

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFv2APIKeyResource,
		Scope:  nuke.Account,
		Lister: &WAFv2APIKeyLister{},
	})
}

type WAFv2APIKeyLister struct{}

func (l *WAFv2APIKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &wafv2.ListAPIKeysInput{
		Limit: aws.Int64(50),
		Scope: aws.String("REGIONAL"),
	}

	output, err := getAPIKeys(svc, params)
	if err != nil {
		return []resource.Resource{}, err
	}

	resources = append(resources, output...)

	if *opts.Session.Config.Region == endpoints.UsEast1RegionID {
		params.Scope = aws.String("CLOUDFRONT")

		output, err := getAPIKeys(svc, params)
		if err != nil {
			return []resource.Resource{}, err
		}

		resources = append(resources, output...)
	}

	return resources, nil
}

func getAPIKeys(svc *wafv2.WAFV2, params *wafv2.ListAPIKeysInput) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.ListAPIKeys(params)
		if err != nil {
			return nil, err
		}

		for _, apiKey := range resp.APIKeySummaries {
			var tokenDomains []string
			for _, tokenDomain := range apiKey.TokenDomains {
				tokenDomains = append(tokenDomains, *tokenDomain)
			}
			resources = append(resources, &WAFv2APIKey{
				svc:          svc,
				apiKey:       apiKey.APIKey,
				tokenDomains: tokenDomains,
				scope:        params.Scope,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}
	return resources, nil
}

type WAFv2APIKey struct {
	svc          *wafv2.WAFV2
	apiKey       *string
	tokenDomains []string
	scope        *string
}

func (f *WAFv2APIKey) Remove(_ context.Context) error {
	_, err := f.svc.DeleteAPIKey(&wafv2.DeleteAPIKeyInput{
		APIKey: f.apiKey,
		Scope:  f.scope,
	})

	return err
}

func (f *WAFv2APIKey) String() string {
	return strings.Join(f.tokenDomains, ", ")
}

func (f *WAFv2APIKey) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("tokenDomains", strings.Join(f.tokenDomains, ", "))
	return properties
}
