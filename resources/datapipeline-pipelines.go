package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/datapipeline" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DataPipelinePipelineResource = "DataPipelinePipeline"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataPipelinePipelineResource,
		Scope:    nuke.Account,
		Resource: &DataPipelinePipeline{},
		Lister:   &DataPipelinePipelineLister{},
	})
}

type DataPipelinePipelineLister struct{}

func (l *DataPipelinePipelineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := datapipeline.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &datapipeline.ListPipelinesInput{}

	for {
		resp, err := svc.ListPipelines(params)
		if err != nil {
			return nil, err
		}

		for _, pipeline := range resp.PipelineIdList {
			resources = append(resources, &DataPipelinePipeline{
				svc:        svc,
				pipelineID: pipeline.Id,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type DataPipelinePipeline struct {
	svc        *datapipeline.DataPipeline
	pipelineID *string
}

func (f *DataPipelinePipeline) Remove(_ context.Context) error {
	_, err := f.svc.DeletePipeline(&datapipeline.DeletePipelineInput{
		PipelineId: f.pipelineID,
	})

	return err
}

func (f *DataPipelinePipeline) String() string {
	return *f.pipelineID
}
