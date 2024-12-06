package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueDataBrewProjectsResource = "GlueDataBrewProjects"

func init() {
	registry.Register(&registry.Registration{
		Name:     GlueDataBrewProjectsResource,
		Scope:    nuke.Account,
		Resource: &GlueDataBrewProjects{},
		Lister:   &GlueDataBrewProjectsLister{},
	})
}

type GlueDataBrewProjectsLister struct{}

func (l *GlueDataBrewProjectsLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gluedatabrew.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &gluedatabrew.ListProjectsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListProjects(params)
		if err != nil {
			return nil, err
		}

		for _, project := range output.Projects {
			resources = append(resources, &GlueDataBrewProjects{
				svc:  svc,
				name: project.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDataBrewProjects struct {
	svc  *gluedatabrew.GlueDataBrew
	name *string
}

func (f *GlueDataBrewProjects) Remove(_ context.Context) error {
	_, err := f.svc.DeleteProject(&gluedatabrew.DeleteProjectInput{
		Name: f.name,
	})

	return err
}

func (f *GlueDataBrewProjects) String() string {
	return *f.name
}
