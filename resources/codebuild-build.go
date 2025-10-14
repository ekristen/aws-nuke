package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codebuild" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeBuildBuildResource = "CodeBuildBuild"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeBuildBuildResource,
		Scope:    nuke.Account,
		Resource: &CodeBuildBuild{},
		Lister:   &CodeBuildBuildLister{},
	})
}

type CodeBuildBuildLister struct{}

func (l *CodeBuildBuildLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codebuild.New(opts.Session)

	params := &codebuild.ListBuildsInput{}

	for {
		resp, err := svc.ListBuilds(params)
		if err != nil {
			return nil, err
		}

		for _, buildID := range resp.Ids {
			resources = append(resources, &CodeBuildBuild{
				svc: svc,
				ID:  buildID,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeBuildBuild struct {
	svc *codebuild.CodeBuild
	ID  *string
}

func (r *CodeBuildBuild) Remove(_ context.Context) error {
	_, err := r.svc.BatchDeleteBuilds(&codebuild.BatchDeleteBuildsInput{
		Ids: []*string{r.ID},
	})

	return err
}

func (r *CodeBuildBuild) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeBuildBuild) String() string {
	return *r.ID
}
