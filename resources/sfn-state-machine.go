package resources

import (
	"context"
	"github.com/ekristen/libnuke/pkg/types"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SFNStateMachineResource = "SFNStateMachine"

func init() {
	registry.Register(&registry.Registration{
		Name:     SFNStateMachineResource,
		Scope:    nuke.Account,
		Resource: &SFNStateMachine{},
		Lister:   &SFNStateMachineLister{},
	})
}

type SFNStateMachineLister struct{}

func (l *SFNStateMachineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sfn.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sfn.ListStateMachinesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListStateMachines(params)
		if err != nil {
			return nil, err
		}

		for _, stateMachine := range output.StateMachines {
			var resourceTags []*sfn.Tag
			tags, err := svc.ListTagsForResource(&sfn.ListTagsForResourceInput{
				ResourceArn: stateMachine.StateMachineArn,
			})
			if err != nil {
				opts.Logger.WithError(err).Error("unable to list state machine tags")
			} else {
				resourceTags = tags.Tags
			}

			resources = append(resources, &SFNStateMachine{
				svc:          svc,
				ARN:          stateMachine.StateMachineArn,
				Name:         stateMachine.Name,
				Type:         stateMachine.Type,
				CreationDate: stateMachine.CreationDate,
				Tags:         resourceTags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SFNStateMachine struct {
	svc          *sfn.SFN
	ARN          *string    `description:"The Amazon Resource Name (ARN) that identifies the state machine."`
	Name         *string    `description:"The name of the state machine."`
	Type         *string    `description:"The type of the state machine."`
	CreationDate *time.Time `description:"The date the state machine was created."`
	Tags         []*sfn.Tag `description:"The tags associated with the state machine."`
}

func (r *SFNStateMachine) Remove(_ context.Context) error {
	_, err := r.svc.DeleteStateMachine(&sfn.DeleteStateMachineInput{
		StateMachineArn: r.ARN,
	})

	return err
}

func (r *SFNStateMachine) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SFNStateMachine) String() string {
	return *r.ARN
}
