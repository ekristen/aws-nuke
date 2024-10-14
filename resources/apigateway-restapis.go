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

const APIGatewayRestAPIResource = "APIGatewayRestAPI"

// Rate limit to avoid throttling when deleting API Gateway Rest APIs
// The API Gateway Delete Rest API has a limit of 1 request per 30 seconds for each account
// https://docs.aws.amazon.com/apigateway/latest/developerguide/limits.html
// Note: due to time drift, set to 31 seconds to be safe.
var deleteRestAPILimit = ratelimit.New(1, ratelimit.Per(32*time.Second))

func init() {
	registry.Register(&registry.Registration{
		Name:   APIGatewayRestAPIResource,
		Scope:  nuke.Account,
		Lister: &APIGatewayRestAPILister{},
	})
}

type APIGatewayRestAPILister struct{}

type APIGatewayRestAPI struct {
	svc         *apigateway.APIGateway
	restAPIID   *string
	name        *string
	version     *string
	createdDate *time.Time
	tags        map[string]*string
}

func (l *APIGatewayRestAPILister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := apigateway.New(opts.Session)

	var resources []resource.Resource

	params := &apigateway.GetRestApisInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.GetRestApis(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Items {
			resources = append(resources, &APIGatewayRestAPI{
				svc:         svc,
				restAPIID:   item.Id,
				name:        item.Name,
				version:     item.Version,
				createdDate: item.CreatedDate,
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

func (f *APIGatewayRestAPI) Remove(_ context.Context) error {
	deleteRestAPILimit.Take()

	_, err := f.svc.DeleteRestApi(&apigateway.DeleteRestApiInput{
		RestApiId: f.restAPIID,
	})

	return err
}

func (f *APIGatewayRestAPI) String() string {
	return *f.restAPIID
}

func (f *APIGatewayRestAPI) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range f.tags {
		properties.SetTag(&key, tag)
	}
	properties.
		Set("APIID", f.restAPIID).
		Set("Name", f.name).
		Set("Version", f.version).
		Set("CreatedDate", f.createdDate.Format(time.RFC3339))
	return properties
}
