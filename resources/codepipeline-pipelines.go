package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codepipeline"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodePipelinePipelineResource = "CodePipelinePipeline"

func init() {
	resource.Register(resource.Registration{
		Name:   CodePipelinePipelineResource,
		Scope:  nuke.Account,
		Lister: &CodePipelinePipelineLister{},
	})
}

type CodePipelinePipelineLister struct{}

func (l *CodePipelinePipelineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codepipeline.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codepipeline.ListPipelinesInput{}

	for {
		resp, err := svc.ListPipelines(params)
		if err != nil {
			return nil, err
		}

		for _, pipeline := range resp.Pipelines {
			resources = append(resources, &CodePipelinePipeline{
				svc:          svc,
				pipelineName: pipeline.Name,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodePipelinePipeline struct {
	svc          *codepipeline.CodePipeline
	pipelineName *string
}

func (f *CodePipelinePipeline) Remove(_ context.Context) error {
	_, err := f.svc.DeletePipeline(&codepipeline.DeletePipelineInput{
		Name: f.pipelineName,
	})

	return err
}

func (f *CodePipelinePipeline) String() string {
	return *f.pipelineName
}
