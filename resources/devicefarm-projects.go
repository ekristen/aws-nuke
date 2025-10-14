package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/devicefarm" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DeviceFarmProjectResource = "DeviceFarmProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     DeviceFarmProjectResource,
		Scope:    nuke.Account,
		Resource: &DeviceFarmProject{},
		Lister:   &DeviceFarmProjectLister{},
	})
}

type DeviceFarmProjectLister struct{}

func (l *DeviceFarmProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := devicefarm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &devicefarm.ListProjectsInput{}

	for {
		output, err := svc.ListProjects(params)
		if err != nil {
			return nil, err
		}

		for _, project := range output.Projects {
			resources = append(resources, &DeviceFarmProject{
				svc: svc,
				ARN: project.Arn,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type DeviceFarmProject struct {
	svc *devicefarm.DeviceFarm
	ARN *string
}

func (f *DeviceFarmProject) Remove(_ context.Context) error {
	_, err := f.svc.DeleteProject(&devicefarm.DeleteProjectInput{
		Arn: f.ARN,
	})

	return err
}

func (f *DeviceFarmProject) String() string {
	return *f.ARN
}
