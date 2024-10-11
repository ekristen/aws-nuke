package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/osis"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OSPipelineResource = "OSPipeline"

func init() {
	registry.Register(&registry.Registration{
		Name:   OSPipelineResource,
		Scope:  nuke.Account,
		Lister: &OSPipelineLister{},
	})
}

type OSPipelineLister struct{}

func (l *OSPipelineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := osis.New(opts.Session)

	params := &osis.ListPipelinesInput{}

	for {
		res, err := svc.ListPipelines(params)
		if err != nil {
			return nil, err
		}

		for _, p := range res.Pipelines {
			resources = append(resources, &OSPipeline{
				svc:       svc,
				Name:      p.PipelineName,
				Tags:      p.Tags,
				Status:    p.Status,
				CreatedAt: p.CreatedAt,
			})
		}

		if res.NextToken == nil {
			break
		}

		params.NextToken = res.NextToken
	}

	return resources, nil
}

type OSPipeline struct {
	svc       *osis.OSIS
	Name      *string
	Status    *string
	CreatedAt *time.Time
	Tags      []*osis.Tag
}

func (r *OSPipeline) Remove(_ context.Context) error {
	_, err := r.svc.DeletePipeline(&osis.DeletePipelineInput{
		PipelineName: r.Name,
	})
	return err
}

func (r *OSPipeline) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *OSPipeline) String() string {
	return *r.Name
}
