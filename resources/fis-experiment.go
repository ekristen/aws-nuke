package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	fistypes "github.com/aws/aws-sdk-go-v2/service/fis/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FISExperimentResource = "FISExperiment"

func init() {
	registry.Register(&registry.Registration{
		Name:     FISExperimentResource,
		Scope:    nuke.Account,
		Resource: &FISExperiment{},
		Lister:   &FISExperimentLister{},
	})
}

type FISExperimentLister struct {
	svc FISAPI
}

func (l *FISExperimentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = fis.NewFromConfig(*opts.Config)
	}

	var resources []resource.Resource

	params := &fis.ListExperimentsInput{
		MaxResults: aws.Int32(100),
	}

	for {
		resp, err := l.svc.ListExperiments(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Experiments {
			experiment := &FISExperiment{
				svc:                  l.svc,
				ID:                   item.Id,
				ARN:                  item.Arn,
				ExperimentTemplateID: item.ExperimentTemplateId,
				CreationTime:         item.CreationTime,
				Tags:                 item.Tags,
			}

			if item.State != nil {
				experiment.Status = string(item.State.Status)
			}

			resources = append(resources, experiment)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type FISExperiment struct {
	svc FISAPI

	ID                   *string           `description:"The ID of the experiment"`
	ARN                  *string           `description:"The ARN of the experiment"`
	ExperimentTemplateID *string           `description:"The ID of the experiment template used to start the experiment"`
	Status               string            `description:"The state of the experiment"`
	CreationTime         *time.Time        `description:"The time that the experiment was created"`
	Tags                 map[string]string `description:"The tags associated with the experiment"`
}

func (r *FISExperiment) Filter() error {
	switch fistypes.ExperimentStatus(r.Status) {
	case fistypes.ExperimentStatusCompleted,
		fistypes.ExperimentStatusStopped,
		fistypes.ExperimentStatusFailed,
		fistypes.ExperimentStatusCancelled,
		fistypes.ExperimentStatusStopping:
		return fmt.Errorf("already %s", r.Status)
	}

	return nil
}

func (r *FISExperiment) Remove(ctx context.Context) error {
	_, err := r.svc.StopExperiment(ctx, &fis.StopExperimentInput{
		Id: r.ID,
	})
	return err
}

func (r *FISExperiment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *FISExperiment) String() string {
	return *r.ID
}
