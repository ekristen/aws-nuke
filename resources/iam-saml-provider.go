package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
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
			svc:        svc,
			ARN:        out.Arn,
			CreateDate: out.CreateDate,
		})
	}

	return resources, nil
}

type IAMSAMLProvider struct {
	svc        iamiface.IAMAPI
	ARN        *string
	CreateDate *time.Time
}

func (r *IAMSAMLProvider) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSAMLProvider(&iam.DeleteSAMLProviderInput{
		SAMLProviderArn: r.ARN,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *IAMSAMLProvider) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IAMSAMLProvider) String() string {
	return *r.ARN
}
