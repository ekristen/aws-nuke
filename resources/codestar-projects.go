package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestar"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeStarProjectResource = "CodeStarProject"

func init() {
	resource.Register(resource.Registration{
		Name:   CodeStarProjectResource,
		Scope:  nuke.Account,
		Lister: &CodeStarProjectLister{},
	})
}

type CodeStarProjectLister struct{}

func (l *CodeStarProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codestar.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codestar.ListProjectsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListProjects(params)
		if err != nil {
			return nil, err
		}

		for _, project := range output.Projects {
			resources = append(resources, &CodeStarProject{
				svc: svc,
				id:  project.ProjectId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CodeStarProject struct {
	svc *codestar.CodeStar
	id  *string
}

func (f *CodeStarProject) Remove(_ context.Context) error {
	_, err := f.svc.DeleteProject(&codestar.DeleteProjectInput{
		Id: f.id,
	})

	return err
}

func (f *CodeStarProject) String() string {
	return *f.id
}
