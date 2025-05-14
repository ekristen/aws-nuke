package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrockagent"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockFlowAliasResource = "BedrockFlowAlias"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockFlowAliasResource,
		Scope:    nuke.Account,
		Lister:   &BedrockFlowAliasLister{},
		Resource: &BedrockFlowAlias{},
	})
}

type BedrockFlowAliasLister struct{}

func (l *BedrockFlowAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	flowIDs, err := ListBedrockFlowIds(svc)
	if err != nil {
		return nil, err
	}

	for _, flowID := range flowIDs {
		params := &bedrockagent.ListFlowAliasesInput{
			MaxResults:     aws.Int64(100),
			FlowIdentifier: aws.String(flowID),
		}
		for {
			output, err := svc.ListFlowAliases(params)
			if err != nil {
				return nil, err
			}

			for _, flowAliasInfo := range output.FlowAliasSummaries {
				resources = append(resources, &BedrockFlowAlias{
					svc:           svc,
					FlowID:        flowAliasInfo.FlowId,
					FlowAliasID:   flowAliasInfo.Id,
					FlowAliasName: flowAliasInfo.Name,
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

type BedrockFlowAlias struct {
	svc           *bedrockagent.BedrockAgent
	FlowID        *string
	FlowAliasID   *string
	FlowAliasName *string
}

func (r *BedrockFlowAlias) Filter() error {
	if strings.HasPrefix(*r.FlowAliasName, "TSTALIASID") {
		return fmt.Errorf("cannot delete AWS managed Flow Alias")
	}
	return nil
}

func (r *BedrockFlowAlias) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockFlowAlias) Remove(_ context.Context) error {
	_, err := r.svc.DeleteFlowAlias(&bedrockagent.DeleteFlowAliasInput{
		AliasIdentifier: r.FlowAliasID,
		FlowIdentifier:  r.FlowID,
	})
	return err
}

func (r *BedrockFlowAlias) String() string {
	return *r.FlowAliasName
}

func ListBedrockFlowIds(svc *bedrockagent.BedrockAgent) ([]string, error) {
	flowIds := []string{}
	params := &bedrockagent.ListFlowsInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListFlows(params)
		if err != nil {
			return nil, err
		}

		for _, flow := range output.FlowSummaries {
			flowIds = append(flowIds, *flow.Id)
		}

		if output.NextToken == nil {
			break
		}
		params.NextToken = output.NextToken
	}

	return flowIds, nil
}
