package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mobile"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MobileProjectResource = "MobileProject"

func init() {
	resource.Register(resource.Registration{
		Name:   MobileProjectResource,
		Scope:  nuke.Account,
		Lister: &MobileProjectLister{},
	})
}

type MobileProjectLister struct{}

func (l *MobileProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mobile.New(opts.Session)
	svc.ClientInfo.SigningName = "AWSMobileHubService"
	resources := make([]resource.Resource, 0)

	params := &mobile.ListProjectsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListProjects(params)
		if err != nil {
			return nil, err
		}

		for _, project := range output.Projects {
			resources = append(resources, &MobileProject{
				svc:       svc,
				projectID: project.ProjectId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MobileProject struct {
	svc       *mobile.Mobile
	projectID *string
}

func (f *MobileProject) Remove(_ context.Context) error {
	_, err := f.svc.DeleteProject(&mobile.DeleteProjectInput{
		ProjectId: f.projectID,
	})

	return err
}

func (f *MobileProject) String() string {
	return *f.projectID
}
