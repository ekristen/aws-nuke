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

const BedrockAgentCoreCodeInterpreterResource = "BedrockAgentCoreCodeInterpreter"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreCodeInterpreterResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreCodeInterpreter{},
		Lister:   &BedrockAgentCoreCodeInterpreterLister{},
	})
}

type BedrockAgentCoreCodeInterpreterLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreCodeInterpreterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	l.SetSupportedRegions(BuiltInToolsSupportedRegions)

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

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
			resources = append(resources, &BedrockAgentCoreCodeInterpreter{
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

type BedrockAgentCoreCodeInterpreter struct {
	svc                *bedrockagentcorecontrol.Client
	CodeInterpreterID  *string
	CodeInterpreterArn *string
	Name               *string
	Status             string
	Description        *string
	CreatedAt          *time.Time
	LastUpdatedAt      *time.Time
}

func (r *BedrockAgentCoreCodeInterpreter) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteCodeInterpreter(ctx, &bedrockagentcorecontrol.DeleteCodeInterpreterInput{
		CodeInterpreterId: r.CodeInterpreterID,
	})

	return err
}

func (r *BedrockAgentCoreCodeInterpreter) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreCodeInterpreter) String() string {
	return *r.CodeInterpreterID
}
