package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueDataBrewJobsResource = "GlueDataBrewJobs"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueDataBrewJobsResource,
		Scope:  nuke.Account,
		Lister: &GlueDataBrewJobsLister{},
	})
}

type GlueDataBrewJobsLister struct{}

func (l *GlueDataBrewJobsLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gluedatabrew.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &gluedatabrew.ListJobsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListJobs(params)
		if err != nil {
			return nil, err
		}

		for _, jobs := range output.Jobs {
			resources = append(resources, &GlueDataBrewJobs{
				svc:  svc,
				name: jobs.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDataBrewJobs struct {
	svc  *gluedatabrew.GlueDataBrew
	name *string
}

func (f *GlueDataBrewJobs) Remove(_ context.Context) error {
	_, err := f.svc.DeleteJob(&gluedatabrew.DeleteJobInput{
		Name: f.name,
	})

	return err
}

func (f *GlueDataBrewJobs) String() string {
	return *f.name
}
