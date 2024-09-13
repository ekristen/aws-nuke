package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/pipes"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const PipesPipeResource = "PipesPipe"

func init() {
	registry.Register(&registry.Registration{
		Name:   PipesPipeResource,
		Scope:  nuke.Account,
		Lister: &PipesPipeLister{},
	})
}

type PipesPipeLister struct{}

func (l *PipesPipeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := pipes.New(opts.Session)
	var resources []resource.Resource

	res, err := svc.ListPipes(&pipes.ListPipesInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.Pipes {
		tagResp, tagsErr := svc.ListTagsForResource(&pipes.ListTagsForResourceInput{
			ResourceArn: p.Arn,
		})

		if tagsErr != nil {
			logrus.WithError(tagsErr).Error("unable to get tags for pipe")
		}

		resources = append(resources, &PipesPipes{
			svc:          svc,
			Name:         p.Name,
			CurrentState: p.CurrentState,
			Source:       p.Source,
			Target:       p.Target,
			CreationDate: p.CreationTime,
			ModifiedDate: p.LastModifiedTime,
			Tags:         tagResp.Tags,
		})
	}

	return resources, nil
}

type PipesPipes struct {
	svc          *pipes.Pipes
	Name         *string
	CurrentState *string
	Source       *string
	Target       *string
	CreationDate *time.Time
	ModifiedDate *time.Time
	Tags         map[string]*string
}

func (r *PipesPipes) Remove(_ context.Context) error {
	_, err := r.svc.DeletePipe(&pipes.DeletePipeInput{
		Name: r.Name,
	})
	return err
}

func (r *PipesPipes) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PipesPipes) String() string {
	return *r.Name
}
