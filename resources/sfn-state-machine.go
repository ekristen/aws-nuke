package resources

import (
	"context"

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
			resources = append(resources, &SFNStateMachine{
				svc: svc,
				ARN: stateMachine.StateMachineArn,
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
	svc *sfn.SFN
	ARN *string
}

func (r *SFNStateMachine) Remove(_ context.Context) error {
	_, err := r.svc.DeleteStateMachine(&sfn.DeleteStateMachineInput{
		StateMachineArn: r.ARN,
	})

	return err
}

func (r *SFNStateMachine) String() string {
	return *r.ARN
}
