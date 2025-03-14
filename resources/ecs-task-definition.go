package resources

import (
	"context"
	"errors"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ECSTaskDefinitionResource = "ECSTaskDefinition"

func init() {
	registry.Register(&registry.Registration{
		Name:     ECSTaskDefinitionResource,
		Scope:    nuke.Account,
		Resource: &ECSTaskDefinition{},
		Lister:   &ECSTaskDefinitionLister{},
	})
}

type ECSTaskDefinitionLister struct{}

func (l *ECSTaskDefinitionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ecs.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	statuses := []ecstypes.TaskDefinitionStatus{
		ecstypes.TaskDefinitionStatusActive,
		ecstypes.TaskDefinitionStatusInactive,
		ecstypes.TaskDefinitionStatusDeleteInProgress,
	}

	for _, status := range statuses {
		params := &ecs.ListTaskDefinitionsInput{
			MaxResults: ptr.Int32(100),
			Status:     status,
		}

		for {
			output, err := svc.ListTaskDefinitions(ctx, params)
			if err != nil {
				var errSkipRequest = liberrors.ErrSkipRequest("skip global")
				if errors.As(err, &errSkipRequest) {
					break
				}

				opts.Logger.Error("unable to list task definitions", "error", err)
				break
			}

			for _, taskDefinitionARN := range output.TaskDefinitionArns {
				details, err := svc.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: ptr.String(taskDefinitionARN),
				})
				if err != nil {
					opts.Logger.Error("unable to describe task definition", "error", err)
					continue
				}

				resources = append(resources, &ECSTaskDefinition{
					svc:    svc,
					arn:    ptr.String(taskDefinitionARN),
					Name:   details.TaskDefinition.Family,
					Status: ptr.String(string(details.TaskDefinition.Status)),
				})
			}

			if output.NextToken == nil {
				break
			}

			params.NextToken = output.NextToken
		}
	}

	return resources, nil
}

type ECSTaskDefinition struct {
	svc    *ecs.Client
	arn    *string
	Name   *string
	Status *string
}

func (r *ECSTaskDefinition) Filter() error {
	if *r.Status == string(ecstypes.TaskDefinitionStatusDeleteInProgress) {
		return errors.New("task definition is in delete in progress status")
	}

	return nil
}

func (r *ECSTaskDefinition) Remove(ctx context.Context) error {
	_, err := r.svc.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: r.arn,
	})

	return err
}

func (r *ECSTaskDefinition) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ECSTaskDefinition) String() string {
	return *r.arn
}
