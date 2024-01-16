package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elastictranscoder"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ElasticTranscoderPipelineResource = "ElasticTranscoderPipeline"

func init() {
	resource.Register(resource.Registration{
		Name:   ElasticTranscoderPipelineResource,
		Scope:  nuke.Account,
		Lister: &ElasticTranscoderPipelineLister{},
	})
}

type ElasticTranscoderPipelineLister struct{}

func (l *ElasticTranscoderPipelineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elastictranscoder.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &elastictranscoder.ListPipelinesInput{}

	for {
		resp, err := svc.ListPipelines(params)
		if err != nil {
			return nil, err
		}

		for _, pipeline := range resp.Pipelines {
			resources = append(resources, &ElasticTranscoderPipeline{
				svc:        svc,
				pipelineID: pipeline.Id,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ElasticTranscoderPipeline struct {
	svc        *elastictranscoder.ElasticTranscoder
	pipelineID *string
}

func (f *ElasticTranscoderPipeline) Remove(_ context.Context) error {
	_, err := f.svc.DeletePipeline(&elastictranscoder.DeletePipelineInput{
		Id: f.pipelineID,
	})

	return err
}

func (f *ElasticTranscoderPipeline) String() string {
	return *f.pipelineID
}
