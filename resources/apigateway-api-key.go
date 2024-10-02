package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayAPIKeyResource = "APIGatewayAPIKey"

func init() {
	registry.Register(&registry.Registration{
		Name:                APIGatewayAPIKeyResource,
		Scope:               nuke.Account,
		Lister:              &APIGatewayAPIKeyLister{},
		AlternativeResource: "AWS::ApiGateway::ApiKey",
	})
}

type APIGatewayAPIKeyLister struct{}

func (l *APIGatewayAPIKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.New(opts.Session)

	var resources []resource.Resource

	params := &apigateway.GetApiKeysInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.GetApiKeys(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayAPIKey{
				svc:         svc,
				apiKey:      item.Id,
				Name:        item.Name,
				Tags:        item.Tags,
				CreatedDate: item.CreatedDate,
			})
		}

		if output.Position == nil {
			break
		}

		params.Position = output.Position
	}

	return resources, nil
}

type APIGatewayAPIKey struct {
	svc         *apigateway.APIGateway
	apiKey      *string
	Name        *string
	Tags        map[string]*string
	CreatedDate *time.Time
}

func (r *APIGatewayAPIKey) Remove(_ context.Context) error {
	_, err := r.svc.DeleteApiKey(&apigateway.DeleteApiKeyInput{
		ApiKey: r.apiKey,
	})

	return err
}

func (r *APIGatewayAPIKey) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *APIGatewayAPIKey) String() string {
	return *r.apiKey
}
