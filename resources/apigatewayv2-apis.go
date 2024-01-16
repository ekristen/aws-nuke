package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const APIGatewayV2APIResource = "APIGatewayV2API"

func init() {
	resource.Register(resource.Registration{
		Name:   APIGatewayV2APIResource,
		Scope:  nuke.Account,
		Lister: &APIGatewayV2APILister{},
	})
}

type APIGatewayV2APILister struct{}

func (l *APIGatewayV2APILister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigatewayv2.New(opts.Session)
	var resources []resource.Resource

	params := &apigatewayv2.GetApisInput{
		MaxResults: aws.String("100"),
	}

	for {
		output, err := svc.GetApis(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayV2API{
				svc:          svc,
				v2APIID:      item.ApiId,
				name:         item.Name,
				protocolType: item.ProtocolType,
				version:      item.Version,
				tags:         item.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type APIGatewayV2API struct {
	svc          *apigatewayv2.ApiGatewayV2
	v2APIID      *string
	name         *string
	protocolType *string
	version      *string
	tags         map[string]*string
}

func (f *APIGatewayV2API) Remove(_ context.Context) error {
	_, err := f.svc.DeleteApi(&apigatewayv2.DeleteApiInput{
		ApiId: f.v2APIID,
	})

	return err
}

func (f *APIGatewayV2API) String() string {
	return *f.v2APIID
}

func (f *APIGatewayV2API) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range f.tags {
		properties.SetTag(&key, tag)
	}
	properties.
		Set("APIID", f.v2APIID).
		Set("Name", f.name).
		Set("ProtocolType", f.protocolType).
		Set("Version", f.version)
	return properties
}
