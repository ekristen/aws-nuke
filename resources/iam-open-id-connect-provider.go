package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IAMOpenIDConnectProviderResource = "IAMOpenIDConnectProvider"

func init() {
	resource.Register(resource.Registration{
		Name:   IAMOpenIDConnectProviderResource,
		Scope:  nuke.Account,
		Lister: &IAMOpenIDConnectProviderLister{},
	})
}

type IAMOpenIDConnectProviderLister struct{}

func (l *IAMOpenIDConnectProviderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	listParams := &iam.ListOpenIDConnectProvidersInput{}
	resources := make([]resource.Resource, 0)

	resp, err := svc.ListOpenIDConnectProviders(listParams)
	if err != nil {
		return nil, err
	}

	for _, out := range resp.OpenIDConnectProviderList {
		params := &iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: out.Arn,
		}
		resp, err := svc.GetOpenIDConnectProvider(params)

		if err != nil {
			return nil, err
		}

		resources = append(resources, &IAMOpenIDConnectProvider{
			svc:  svc,
			arn:  *out.Arn,
			tags: resp.Tags,
		})
	}

	return resources, nil
}

type IAMOpenIDConnectProvider struct {
	svc  iamiface.IAMAPI
	arn  string
	tags []*iam.Tag
}

func (e *IAMOpenIDConnectProvider) Remove(_ context.Context) error {
	_, err := e.svc.DeleteOpenIDConnectProvider(&iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMOpenIDConnectProvider) String() string {
	return e.arn
}

func (e *IAMOpenIDConnectProvider) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Arn", e.arn)

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
