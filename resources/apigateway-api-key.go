package resources

import (
	"context"
	"time"

	"go.uber.org/ratelimit"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const APIGatewayAPIKeyResource = "APIGatewayAPIKey"

// Rate limit to avoid throttling when deleting API Gateway Rest APIs
// The API Gateway Delete Rest API has a limit of 1 request per 30 seconds for each account
// https://docs.aws.amazon.com/apigateway/latest/developerguide/limits.html
// Note: due to time drift, set to 31 seconds to be safe.
var deleteAPIKeyLimit = ratelimit.New(1, ratelimit.Per(32*time.Second))

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
	deleteAPIKeyLimit.Take()

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
