package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SSMMaintenanceWindowResource = "SSMMaintenanceWindow"

func init() {
	registry.Register(&registry.Registration{
		Name:   SSMMaintenanceWindowResource,
		Scope:  nuke.Account,
		Lister: &SSMMaintenanceWindowLister{},
	})
}

type SSMMaintenanceWindowLister struct{}

func (l *SSMMaintenanceWindowLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ssm.DescribeMaintenanceWindowsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeMaintenanceWindows(params)
		if err != nil {
			return nil, err
		}

		for _, windowIdentity := range output.WindowIdentities {
			resources = append(resources, &SSMMaintenanceWindow{
				svc: svc,
				ID:  windowIdentity.WindowId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMMaintenanceWindow struct {
	svc *ssm.SSM
	ID  *string
}

func (f *SSMMaintenanceWindow) Remove(_ context.Context) error {
	_, err := f.svc.DeleteMaintenanceWindow(&ssm.DeleteMaintenanceWindowInput{
		WindowId: f.ID,
	})

	return err
}

func (f *SSMMaintenanceWindow) String() string {
	return *f.ID
}
