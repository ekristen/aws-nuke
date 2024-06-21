package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ECSTaskDefinitionResource = "ECSTaskDefinition"

func init() {
	registry.Register(&registry.Registration{
		Name:   ECSTaskDefinitionResource,
		Scope:  nuke.Account,
		Lister: &ECSTaskDefinitionLister{},
	})
}

type ECSTaskDefinitionLister struct{}

func (l *ECSTaskDefinitionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ecs.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ecs.ListTaskDefinitionsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListTaskDefinitions(params)
		if err != nil {
			return nil, err
		}

		for _, taskDefinitionARN := range output.TaskDefinitionArns {
			resources = append(resources, &ECSTaskDefinition{
				svc: svc,
				ARN: taskDefinitionARN,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSTaskDefinition struct {
	svc *ecs.ECS
	ARN *string
}

func (f *ECSTaskDefinition) Remove(_ context.Context) error {
	_, err := f.svc.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: f.ARN,
	})

	return err
}

func (f *ECSTaskDefinition) String() string {
	return *f.ARN
}
