package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlobalAcceleratorEndpointGroupResource = "GlobalAcceleratorEndpointGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlobalAcceleratorEndpointGroupResource,
		Scope:  nuke.Account,
		Lister: &GlobalAcceleratorEndpointGroupLister{},
	})
}

type GlobalAcceleratorEndpointGroupLister struct{}

// List enumerates all available accelerators
func (l *GlobalAcceleratorEndpointGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := globalaccelerator.New(opts.Session)
	var acceleratorARNs []*string
	var listenerARNs []*string
	resources := make([]resource.Resource, 0)

	// get all accelerator ARNs
	acceleratorParams := &globalaccelerator.ListAcceleratorsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListAccelerators(acceleratorParams)
		if err != nil {
			return nil, err
		}

		for _, accelerator := range output.Accelerators {
			acceleratorARNs = append(acceleratorARNs, accelerator.AcceleratorArn)
		}

		if output.NextToken == nil {
			break
		}

		acceleratorParams.NextToken = output.NextToken
	}

	// get all listeners ARNs of all accelerators
	for _, acceleratorARN := range acceleratorARNs {
		listenerParams := &globalaccelerator.ListListenersInput{
			MaxResults:     aws.Int64(100),
			AcceleratorArn: acceleratorARN,
		}

		for {
			output, err := svc.ListListeners(listenerParams)
			if err != nil {
				return nil, err
			}

			for _, listener := range output.Listeners {
				listenerARNs = append(listenerARNs, listener.ListenerArn)
			}

			if output.NextToken == nil {
				break
			}

			listenerParams.NextToken = output.NextToken
		}
	}

	// get all endpoints based on all listeners based on all accelerator
	for _, listenerArn := range listenerARNs {
		params := &globalaccelerator.ListEndpointGroupsInput{
			MaxResults:  aws.Int64(100),
			ListenerArn: listenerArn,
		}

		for {
			output, err := svc.ListEndpointGroups(params)
			if err != nil {
				return nil, err
			}

			for _, endpointGroup := range output.EndpointGroups {
				resources = append(resources, &GlobalAcceleratorEndpointGroup{
					svc: svc,
					ARN: endpointGroup.EndpointGroupArn,
				})
			}

			if output.NextToken == nil {
				break
			}

			params.NextToken = output.NextToken
		}
	}

	return resources, nil
}

// GlobalAcceleratorEndpointGroup model
type GlobalAcceleratorEndpointGroup struct {
	svc *globalaccelerator.GlobalAccelerator
	ARN *string
}

// Remove resource
func (g *GlobalAcceleratorEndpointGroup) Remove(_ context.Context) error {
	_, err := g.svc.DeleteEndpointGroup(&globalaccelerator.DeleteEndpointGroupInput{
		EndpointGroupArn: g.ARN,
	})

	return err
}

// Properties definition
func (g *GlobalAcceleratorEndpointGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", g.ARN)
	return properties
}

// String representation
func (g *GlobalAcceleratorEndpointGroup) String() string {
	return *g.ARN
}
