package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayDomainNameResource = "APIGatewayDomainName"

func init() {
	registry.Register(&registry.Registration{
		Name:     APIGatewayDomainNameResource,
		Scope:    nuke.Account,
		Resource: &APIGatewayDomainName{},
		Lister:   &APIGatewayDomainNameLister{},
	})
}

type APIGatewayDomainNameLister struct{}

func (l *APIGatewayDomainNameLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &apigateway.GetDomainNamesInput{
		Limit: aws.Int32(100),
	}

	for {
		output, err := svc.GetDomainNames(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range output.Items {
			item := &output.Items[i]
			resources = append(resources, &APIGatewayDomainName{
				svc:          svc,
				DomainName:   item.DomainName,
				DomainNameID: item.DomainNameId,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

type APIGatewayDomainName struct {
	svc          *apigateway.Client
	DomainName   *string
	DomainNameID *string
}

func (f *APIGatewayDomainName) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteDomainName(ctx, &apigateway.DeleteDomainNameInput{
		DomainName:   f.DomainName,
		DomainNameId: f.DomainNameID,
	})

	return err
}

func (f *APIGatewayDomainName) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *APIGatewayDomainName) String() string {
	return *f.DomainName
}
