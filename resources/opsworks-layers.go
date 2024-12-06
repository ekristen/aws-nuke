package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opsworks"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OpsWorksLayerResource = "OpsWorksLayer"

func init() {
	registry.Register(&registry.Registration{
		Name:     OpsWorksLayerResource,
		Scope:    nuke.Account,
		Resource: &OpsWorksLayer{},
		Lister:   &OpsWorksLayerLister{},
	})
}

type OpsWorksLayerLister struct{}

func (l *OpsWorksLayerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworks.New(opts.Session)
	resources := make([]resource.Resource, 0)

	stackParams := &opsworks.DescribeStacksInput{}

	resp, err := svc.DescribeStacks(stackParams)
	if err != nil {
		return nil, err
	}

	layerParams := &opsworks.DescribeLayersInput{}

	for _, stack := range resp.Stacks {
		layerParams.StackId = stack.StackId
		output, err := svc.DescribeLayers(layerParams)
		if err != nil {
			return nil, err
		}

		for _, layer := range output.Layers {
			resources = append(resources, &OpsWorksLayer{
				svc: svc,
				ID:  layer.LayerId,
			})
		}
	}

	return resources, nil
}

type OpsWorksLayer struct {
	svc *opsworks.OpsWorks
	ID  *string
}

func (f *OpsWorksLayer) Remove(_ context.Context) error {
	_, err := f.svc.DeleteLayer(&opsworks.DeleteLayerInput{
		LayerId: f.ID,
	})

	return err
}

func (f *OpsWorksLayer) String() string {
	return *f.ID
}
