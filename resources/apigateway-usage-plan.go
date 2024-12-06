package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayUsagePlanResource = "APIGatewayUsagePlan"

func init() {
	registry.Register(&registry.Registration{
		Name:                APIGatewayUsagePlanResource,
		Scope:               nuke.Account,
		Resource:            &APIGatewayUsagePlan{},
		Lister:              &APIGatewayUsagePlanLister{},
		AlternativeResource: "AWS::ApiGateway::UsagePlan",
	})
}

type APIGatewayUsagePlanLister struct{}

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
				UsagePlanID: item.Id,
				Name:        item.Name,
				Tags:        item.Tags,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

type APIGatewayUsagePlan struct {
	svc         *apigateway.APIGateway
	UsagePlanID *string
	Name        *string
	Tags        map[string]*string
}

func (r *APIGatewayUsagePlan) Remove(_ context.Context) error {
	_, err := r.svc.DeleteUsagePlan(&apigateway.DeleteUsagePlanInput{
		UsagePlanId: r.UsagePlanID,
	})

	return err
}

func (r *APIGatewayUsagePlan) String() string {
	return *r.UsagePlanID
}

func (r *APIGatewayUsagePlan) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
