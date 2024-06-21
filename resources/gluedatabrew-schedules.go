package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueDataBrewSchedulesResource = "GlueDataBrewSchedules"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueDataBrewSchedulesResource,
		Scope:  nuke.Account,
		Lister: &GlueDataBrewSchedulesLister{},
	})
}

type GlueDataBrewSchedulesLister struct{}

func (l *GlueDataBrewSchedulesLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gluedatabrew.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &gluedatabrew.ListSchedulesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListSchedules(params)
		if err != nil {
			return nil, err
		}

		for _, schedule := range output.Schedules {
			resources = append(resources, &GlueDataBrewSchedules{
				svc:  svc,
				name: schedule.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDataBrewSchedules struct {
	svc  *gluedatabrew.GlueDataBrew
	name *string
}

func (f *GlueDataBrewSchedules) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSchedule(&gluedatabrew.DeleteScheduleInput{
		Name: f.name,
	})

	return err
}

func (f *GlueDataBrewSchedules) String() string {
	return *f.name
}
