package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

const IAMSAMLProviderResource = "IAMSAMLProvider"

func init() {
	registry.Register(&registry.Registration{
		Name:   IAMSAMLProviderResource,
		Scope:  nuke.Account,
		Lister: &IAMSAMLProviderLister{},
	})
}

type IAMSAMLProviderLister struct{}

func (l *IAMSAMLProviderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)
	params := &iam.ListSAMLProvidersInput{}
	resources := make([]resource.Resource, 0)

	resp, err := svc.ListSAMLProviders(params)
	if err != nil {
		return nil, err
	}

	for _, out := range resp.SAMLProviderList {
		resources = append(resources, &IAMSAMLProvider{
			svc: svc,
			arn: *out.Arn,
		})
	}

	return resources, nil
}

type IAMSAMLProvider struct {
	svc iamiface.IAMAPI
	arn string
}

func (e *IAMSAMLProvider) Remove(_ context.Context) error {
	_, err := e.svc.DeleteSAMLProvider(&iam.DeleteSAMLProviderInput{
		SAMLProviderArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMSAMLProvider) String() string {
	return e.arn
}
