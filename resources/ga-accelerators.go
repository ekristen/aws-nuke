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

// GlobalAccelerator model
type GlobalAccelerator struct {
	svc *globalaccelerator.GlobalAccelerator
	ARN *string
}

const GlobalAcceleratorResource = "GlobalAccelerator"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlobalAcceleratorResource,
		Scope:  nuke.Account,
		Lister: &GlobalAcceleratorLister{},
	})
}

type GlobalAcceleratorLister struct{}

// List enumerates all available accelerators
func (l *GlobalAcceleratorLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := globalaccelerator.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &globalaccelerator.ListAcceleratorsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListAccelerators(params)
		if err != nil {
			return nil, err
		}

		for _, accelerator := range output.Accelerators {
			resources = append(resources, &GlobalAccelerator{
				svc: svc,
				ARN: accelerator.AcceleratorArn,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

// Remove resource
func (ga *GlobalAccelerator) Remove(_ context.Context) error {
	accel, err := ga.svc.DescribeAccelerator(&globalaccelerator.DescribeAcceleratorInput{
		AcceleratorArn: ga.ARN,
	})
	if err != nil {
		return err
	}
	if *accel.Accelerator.Enabled {
		_, err := ga.svc.UpdateAccelerator(&globalaccelerator.UpdateAcceleratorInput{
			AcceleratorArn: ga.ARN,
			Enabled:        aws.Bool(false),
		})
		if err != nil {
			return err
		}
	}
	_, err = ga.svc.DeleteAccelerator(&globalaccelerator.DeleteAcceleratorInput{
		AcceleratorArn: ga.ARN,
	})

	return err
}

// Properties definition
func (ga *GlobalAccelerator) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", ga.ARN)
	return properties
}

// String representation
func (ga *GlobalAccelerator) String() string {
	return *ga.ARN
}
