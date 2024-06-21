package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/robomaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RoboMakerRobotApplicationResource = "RoboMakerRobotApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:   RoboMakerRobotApplicationResource,
		Scope:  nuke.Account,
		Lister: &RoboMakerRobotApplicationLister{},
	})
}

type RoboMakerRobotApplicationLister struct{}

func (l *RoboMakerRobotApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := robomaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &robomaker.ListRobotApplicationsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListRobotApplications(params)
		if err != nil {
			return nil, err
		}

		for _, robotApplication := range resp.RobotApplicationSummaries {
			resources = append(resources, &RoboMakerRobotApplication{
				svc:     svc,
				name:    robotApplication.Name,
				arn:     robotApplication.Arn,
				version: robotApplication.Version,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type RoboMakerRobotApplication struct {
	svc     *robomaker.RoboMaker
	name    *string
	arn     *string
	version *string
}

func (f *RoboMakerRobotApplication) Remove(_ context.Context) error {
	request := robomaker.DeleteRobotApplicationInput{
		Application: f.arn,
	}
	if f.version != nil && *f.version != "$LATEST" {
		request.ApplicationVersion = f.version
	}

	_, err := f.svc.DeleteRobotApplication(&request)

	return err
}

func (f *RoboMakerRobotApplication) String() string {
	return *f.name
}
