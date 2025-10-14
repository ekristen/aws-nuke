package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/apigateway" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayClientCertificateResource = "APIGatewayClientCertificate"

func init() {
	registry.Register(&registry.Registration{
		Name:                APIGatewayClientCertificateResource,
		Scope:               nuke.Account,
		Resource:            &APIGatewayClientCertificate{},
		Lister:              &APIGatewayClientCertificateLister{},
		AlternativeResource: "AWS::ApiGateway::ClientCertificate",
	})
}

type APIGatewayClientCertificateLister struct{}

func (l *APIGatewayClientCertificateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.New(opts.Session)
	var resources []resource.Resource

	params := &apigateway.GetClientCertificatesInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.GetClientCertificates(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayClientCertificate{
				svc:                 svc,
				clientCertificateID: item.ClientCertificateId,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

type APIGatewayClientCertificate struct {
	svc                 *apigateway.APIGateway
	clientCertificateID *string
}

func (f *APIGatewayClientCertificate) Remove(_ context.Context) error {
	_, err := f.svc.DeleteClientCertificate(&apigateway.DeleteClientCertificateInput{
		ClientCertificateId: f.clientCertificateID,
	})

	return err
}

func (f *APIGatewayClientCertificate) String() string {
	return *f.clientCertificateID
}
