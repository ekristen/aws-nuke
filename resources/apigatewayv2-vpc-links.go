package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayV2VpcLinkResource = "APIGatewayV2VpcLink"

func init() {
	registry.Register(&registry.Registration{
		Name:     APIGatewayV2VpcLinkResource,
		Scope:    nuke.Account,
		Resource: &APIGatewayV2VpcLink{},
		Lister:   &APIGatewayV2VpcLinkLister{},
	})
}

type APIGatewayV2VpcLinkLister struct{}

func (l *APIGatewayV2VpcLinkLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigatewayv2.New(opts.Session)
	var resources []resource.Resource

	params := &apigatewayv2.GetVpcLinksInput{
		MaxResults: aws.String("100"),
	}

	for {
		output, err := svc.GetVpcLinks(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayV2VpcLink{
				svc:       svc,
				vpcLinkID: item.VpcLinkId,
				name:      item.Name,
				tags:      item.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type APIGatewayV2VpcLink struct {
	svc       *apigatewayv2.ApiGatewayV2
	vpcLinkID *string
	name      *string
	tags      map[string]*string
}

func (f *APIGatewayV2VpcLink) Remove(_ context.Context) error {
	_, err := f.svc.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: f.vpcLinkID,
	})

	return err
}

func (f *APIGatewayV2VpcLink) String() string {
	return *f.vpcLinkID
}

func (f *APIGatewayV2VpcLink) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range f.tags {
		properties.SetTag(&key, tag)
	}
	properties.
		Set("VPCLinkID", f.vpcLinkID).
		Set("Name", f.name)
	return properties
}
