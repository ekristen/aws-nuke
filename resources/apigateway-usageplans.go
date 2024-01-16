package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const APIGatewayUsagePlanResource = "APIGatewayUsagePlan"

func init() {
	resource.Register(resource.Registration{
		Name:   APIGatewayUsagePlanResource,
		Scope:  nuke.Account,
		Lister: &APIGatewayUsagePlanLister{},
	}, nuke.MapCloudControl("AWS::ApiGateway::UsagePlan"))
}

type APIGatewayUsagePlanLister struct{}

type APIGatewayUsagePlan struct {
	svc         *apigateway.APIGateway
	usagePlanID *string
	name        *string
	tags        map[string]*string
}

func (l *APIGatewayUsagePlanLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.New(opts.Session)
	var resources []resource.Resource

	params := &apigateway.GetUsagePlansInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.GetUsagePlans(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayUsagePlan{
				svc:         svc,
				usagePlanID: item.Id,
				name:        item.Name,
				tags:        item.Tags,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

func (f *APIGatewayUsagePlan) Remove(_ context.Context) error {
	_, err := f.svc.DeleteUsagePlan(&apigateway.DeleteUsagePlanInput{
		UsagePlanId: f.usagePlanID,
	})

	return err
}

func (f *APIGatewayUsagePlan) String() string {
	return *f.usagePlanID
}

func (f *APIGatewayUsagePlan) Properties() types.Properties {
	properties := types.NewProperties()

	for key, tag := range f.tags {
		properties.SetTag(&key, tag)
	}

	properties.
		Set("UsagePlanID", f.usagePlanID).
		Set("Name", f.name)
	return properties
}
