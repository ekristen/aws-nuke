package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codeartifact"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeArtifactRepositoryResource = "CodeArtifactRepository"

func init() {
	resource.Register(&resource.Registration{
		Name:   CodeArtifactRepositoryResource,
		Scope:  nuke.Account,
		Lister: &CodeArtifactRepositoryLister{},
	})
}

type CodeArtifactRepositoryLister struct{}

func (l *CodeArtifactRepositoryLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codeartifact.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codeartifact.ListRepositoriesInput{}

	for {
		resp, err := svc.ListRepositories(params)
		if err != nil {
			return nil, err
		}

		for _, repo := range resp.Repositories {
			resources = append(resources, &CodeArtifactRepository{
				svc:    svc,
				name:   repo.Name,
				domain: repo.DomainName,
				tags:   GetRepositoryTags(svc, repo.Arn),
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func GetRepositoryTags(svc *codeartifact.CodeArtifact, arn *string) map[string]*string {
	tags := map[string]*string{}

	resp, _ := svc.ListTagsForResource(&codeartifact.ListTagsForResourceInput{
		ResourceArn: arn,
	})
	for _, tag := range resp.Tags {
		tags[*tag.Key] = tag.Value
	}

	return tags
}

type CodeArtifactRepository struct {
	svc    *codeartifact.CodeArtifact
	name   *string
	domain *string
	tags   map[string]*string
}

func (r *CodeArtifactRepository) Remove(_ context.Context) error {
	_, err := r.svc.DeleteRepository(&codeartifact.DeleteRepositoryInput{
		Repository: r.name,
		Domain:     r.domain,
	})
	return err
}

func (r *CodeArtifactRepository) String() string {
	return *r.name
}

func (r *CodeArtifactRepository) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range r.tags {
		properties.SetTag(&key, tag)
	}
	properties.Set("Name", r.name)
	properties.Set("Domain", r.domain)
	return properties
}
