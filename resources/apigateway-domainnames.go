package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const APIGatewayDomainNameResource = "APIGatewayDomainName"

func init() {
	registry.Register(&registry.Registration{
		Name:   APIGatewayDomainNameResource,
		Scope:  nuke.Account,
		Lister: &APIGatewayDomainNameLister{},
	})
}

type APIGatewayDomainNameLister struct{}

func (l *APIGatewayDomainNameLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.New(opts.Session)
	var resources []resource.Resource

	params := &apigateway.GetDomainNamesInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.GetDomainNames(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayDomainName{
				svc:        svc,
				domainName: item.DomainName,
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
	svc        *apigateway.APIGateway
	domainName *string
}

func (f *APIGatewayDomainName) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDomainName(&apigateway.DeleteDomainNameInput{
		DomainName: f.domainName,
	})

	return err
}

func (f *APIGatewayDomainName) String() string {
	return *f.domainName
}
