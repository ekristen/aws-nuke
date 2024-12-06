package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeBuildBuildBatchResource = "CodeBuildBuildBatch"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeBuildBuildBatchResource,
		Scope:    nuke.Account,
		Resource: &CodeBuildBuildBatch{},
		Lister:   &CodeBuildBuildBatchLister{},
	})
}

type CodeBuildBuildBatchLister struct{}

func (l *CodeBuildBuildBatchLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codebuild.New(opts.Session)

	params := &codebuild.ListBuildBatchesInput{}

	for {
		resp, err := svc.ListBuildBatches(params)
		if err != nil {
			return nil, err
		}

		for _, batchID := range resp.Ids {
			resources = append(resources, &CodeBuildBuildBatch{
				svc: svc,
				ID:  batchID,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeBuildBuildBatch struct {
	svc *codebuild.CodeBuild
	ID  *string
}

func (r *CodeBuildBuildBatch) Remove(_ context.Context) error {
	_, err := r.svc.DeleteBuildBatch(&codebuild.DeleteBuildBatchInput{
		Id: r.ID,
	})

	return err
}

func (r *CodeBuildBuildBatch) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeBuildBuildBatch) String() string {
	return *r.ID
}
