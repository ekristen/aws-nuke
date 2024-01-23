package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codecommit"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeCommitRepositoryResource = "CodeCommitRepository"

func init() {
	resource.Register(&resource.Registration{
		Name:   CodeCommitRepositoryResource,
		Scope:  nuke.Account,
		Lister: &CodeCommitRepositoryLister{},
	})
}

type CodeCommitRepositoryLister struct{}

// List - Return a list of all CodeCommit Repositories as Resources
func (l *CodeCommitRepositoryLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codecommit.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codecommit.ListRepositoriesInput{}

	for {
		resp, err := svc.ListRepositories(params)
		if err != nil {
			return nil, err
		}

		for _, repository := range resp.Repositories {
			resources = append(resources, &CodeCommitRepository{
				svc:            svc,
				repositoryName: repository.RepositoryName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeCommitRepository struct {
	svc            *codecommit.CodeCommit
	repositoryName *string
}

// Remove - Removes the CodeCommit Repository
func (f *CodeCommitRepository) Remove(_ context.Context) error {
	_, err := f.svc.DeleteRepository(&codecommit.DeleteRepositoryInput{
		RepositoryName: f.repositoryName,
	})

	return err
}

func (f *CodeCommitRepository) String() string {
	return *f.repositoryName
}
