package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const APIGatewayVpcLinkResource = "APIGatewayVpcLink"

func init() {
	resource.Register(resource.Registration{
		Name:   APIGatewayVpcLinkResource,
		Scope:  nuke.Account,
		Lister: &APIGatewayVpcLinkLister{},
	})
}

type APIGatewayVpcLinkLister struct{}

func (l *APIGatewayVpcLinkLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.New(opts.Session)
	var resources []resource.Resource

	params := &apigateway.GetVpcLinksInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.GetVpcLinks(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayVpcLink{
				svc:       svc,
				vpcLinkID: item.Id,
				name:      item.Name,
				tags:      item.Tags,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

type APIGatewayVpcLink struct {
	svc       *apigateway.APIGateway
	vpcLinkID *string
	name      *string
	tags      map[string]*string
}

func (f *APIGatewayVpcLink) Remove(_ context.Context) error {
	_, err := f.svc.DeleteVpcLink(&apigateway.DeleteVpcLinkInput{
		VpcLinkId: f.vpcLinkID,
	})

	return err
}

func (f *APIGatewayVpcLink) String() string {
	return *f.vpcLinkID
}

func (f *APIGatewayVpcLink) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range f.tags {
		properties.SetTag(&key, tag)
	}
	properties.
		Set("VPCLinkID", f.vpcLinkID).
		Set("Name", f.name)
	return properties
}
