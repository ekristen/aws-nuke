package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"               //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/robomaker" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RoboMakerSimulationApplicationResource = "RoboMakerSimulationApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:     RoboMakerSimulationApplicationResource,
		Scope:    nuke.Account,
		Resource: &RoboMakerSimulationApplication{},
		Lister:   &RoboMakerSimulationApplicationLister{},
	})
}

type RoboMakerSimulationApplicationLister struct{}

func (l *RoboMakerSimulationApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := robomaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &robomaker.ListSimulationApplicationsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListSimulationApplications(params)
		if err != nil {
			return nil, err
		}

		for _, robotSimulationApplication := range resp.SimulationApplicationSummaries {
			resources = append(resources, &RoboMakerSimulationApplication{
				svc:     svc,
				name:    robotSimulationApplication.Name,
				arn:     robotSimulationApplication.Arn,
				version: robotSimulationApplication.Version,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type RoboMakerSimulationApplication struct {
	svc     *robomaker.RoboMaker
	name    *string
	arn     *string
	version *string
}

func (f *RoboMakerSimulationApplication) Remove(_ context.Context) error {
	request := robomaker.DeleteSimulationApplicationInput{
		Application: f.arn,
	}
	if f.version != nil && *f.version != "$LATEST" {
		request.ApplicationVersion = f.version
	}
	_, err := f.svc.DeleteSimulationApplication(&request)

	return err
}

func (f *RoboMakerSimulationApplication) String() string {
	return *f.name
}
