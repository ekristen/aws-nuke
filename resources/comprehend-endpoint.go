package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/comprehend"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ComprehendEndpointResource = "ComprehendEndpoint"

func init() {
	resource.Register(resource.Registration{
		Name:   ComprehendEndpointResource,
		Scope:  nuke.Account,
		Lister: &ComprehendEndpointLister{},
	})
}

type ComprehendEndpointLister struct{}

func (l *ComprehendEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := comprehend.New(opts.Session)

	params := &comprehend.ListEndpointsInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListEndpoints(params)
		if err != nil {
			return nil, err
		}
		for _, endpoint := range resp.EndpointPropertiesList {
			resources = append(resources, &ComprehendEndpoint{
				svc:      svc,
				endpoint: endpoint,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type ComprehendEndpoint struct {
	svc      *comprehend.Comprehend
	endpoint *comprehend.EndpointProperties
}

func (ce *ComprehendEndpoint) Remove(_ context.Context) error {
	_, err := ce.svc.DeleteEndpoint(&comprehend.DeleteEndpointInput{
		EndpointArn: ce.endpoint.EndpointArn,
	})
	return err
}

func (ce *ComprehendEndpoint) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("EndpointArn", ce.endpoint.EndpointArn)
	properties.Set("ModelArn", ce.endpoint.ModelArn)

	return properties
}

func (ce *ComprehendEndpoint) String() string {
	return *ce.endpoint.EndpointArn
}
