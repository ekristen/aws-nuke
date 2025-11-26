package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockCodeInterpreterResource = "BedrockCodeInterpreter"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockCodeInterpreterResource,
		Scope:    nuke.Account,
		Resource: &BedrockCodeInterpreter{},
		Lister:   &BedrockCodeInterpreterLister{},
	})
}

type BedrockCodeInterpreterLister struct{}

func (l *BedrockCodeInterpreterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &bedrockagentcorecontrol.ListCodeInterpretersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListCodeInterpretersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, interpreter := range resp.CodeInterpreterSummaries {
			resources = append(resources, &BedrockCodeInterpreter{
				svc:                svc,
				CodeInterpreterID:  interpreter.CodeInterpreterId,
				CodeInterpreterArn: interpreter.CodeInterpreterArn,
				Name:               interpreter.Name,
				Status:             string(interpreter.Status),
				Description:        interpreter.Description,
				CreatedAt:          interpreter.CreatedAt,
				LastUpdatedAt:      interpreter.LastUpdatedAt,
			})
		}
	}

	return resources, nil
}

type BedrockCodeInterpreter struct {
	svc                *bedrockagentcorecontrol.Client
	CodeInterpreterID  *string
	CodeInterpreterArn *string
	Name               *string
	Status             string
	Description        *string
	CreatedAt          *time.Time
	LastUpdatedAt      *time.Time
}

func (r *BedrockCodeInterpreter) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteCodeInterpreter(ctx, &bedrockagentcorecontrol.DeleteCodeInterpreterInput{
		CodeInterpreterId: r.CodeInterpreterID,
	})

	return err
}

func (r *BedrockCodeInterpreter) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockCodeInterpreter) String() string {
	return *r.CodeInterpreterID
}
