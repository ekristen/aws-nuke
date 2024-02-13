package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opsworks"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const OpsWorksInstanceResource = "OpsWorksInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:   OpsWorksInstanceResource,
		Scope:  nuke.Account,
		Lister: &OpsWorksInstanceLister{},
	})
}

type OpsWorksInstanceLister struct{}

func (l *OpsWorksInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworks.New(opts.Session)
	resources := make([]resource.Resource, 0)

	stackParams := &opsworks.DescribeStacksInput{}

	resp, err := svc.DescribeStacks(stackParams)
	if err != nil {
		return nil, err
	}

	instanceParams := &opsworks.DescribeInstancesInput{}
	for _, stack := range resp.Stacks {
		instanceParams.StackId = stack.StackId
		output, err := svc.DescribeInstances(instanceParams)
		if err != nil {
			return nil, err
		}

		for _, instance := range output.Instances {
			resources = append(resources, &OpsWorksInstance{
				svc: svc,
				ID:  instance.InstanceId,
			})
		}
	}

	return resources, nil
}

type OpsWorksInstance struct {
	svc *opsworks.OpsWorks
	ID  *string
}

func (f *OpsWorksInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeleteInstance(&opsworks.DeleteInstanceInput{
		InstanceId: f.ID,
	})

	return err
}

func (f *OpsWorksInstance) String() string {
	return *f.ID
}
