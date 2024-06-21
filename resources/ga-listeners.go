package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlobalAcceleratorListenerResource = "GlobalAcceleratorListener"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlobalAcceleratorListenerResource,
		Scope:  nuke.Account,
		Lister: &GlobalAcceleratorListenerLister{},
	})
}

type GlobalAcceleratorListenerLister struct{}

// List enumerates all available listeners of all available accelerators
func (l *GlobalAcceleratorListenerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := globalaccelerator.New(opts.Session)
	var acceleratorARNs []*string
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

	// get all listeners
	for _, acceleratorARN := range acceleratorARNs {
		params := &globalaccelerator.ListListenersInput{
			MaxResults:     aws.Int64(100),
			AcceleratorArn: acceleratorARN,
		}

		for {
			output, err := svc.ListListeners(params)
			if err != nil {
				return nil, err
			}

			for _, listener := range output.Listeners {
				resources = append(resources, &GlobalAcceleratorListener{
					svc: svc,
					ARN: listener.ListenerArn,
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

// GlobalAcceleratorListener model
type GlobalAcceleratorListener struct {
	svc *globalaccelerator.GlobalAccelerator
	ARN *string
}

// Remove resource
func (g *GlobalAcceleratorListener) Remove(_ context.Context) error {
	_, err := g.svc.DeleteListener(&globalaccelerator.DeleteListenerInput{
		ListenerArn: g.ARN,
	})

	return err
}

// Properties definition
func (g *GlobalAcceleratorListener) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", g.ARN)
	return properties
}

// String representation
func (g *GlobalAcceleratorListener) String() string {
	return *g.ARN
}
