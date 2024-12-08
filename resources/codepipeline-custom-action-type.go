package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/codepipeline"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodePipelineCustomActionTypeResource = "CodePipelineCustomActionType"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodePipelineCustomActionTypeResource,
		Scope:    nuke.Account,
		Resource: &CodePipelineCustomActionType{},
		Lister:   &CodePipelineCustomActionTypeLister{},
	})
}

type CodePipelineCustomActionTypeLister struct{}

func (l *CodePipelineCustomActionTypeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codepipeline.New(opts.Session)

	params := &codepipeline.ListActionTypesInput{}

	for {
		resp, err := svc.ListActionTypes(params)
		if err != nil {
			return nil, err
		}

		for _, actionTypes := range resp.ActionTypes {
			resources = append(resources, &CodePipelineCustomActionType{
				svc:      svc,
				Owner:    actionTypes.Id.Owner,
				Category: actionTypes.Id.Category,
				Provider: actionTypes.Id.Provider,
				Version:  actionTypes.Id.Version,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodePipelineCustomActionType struct {
	svc      *codepipeline.CodePipeline
	Owner    *string
	Category *string
	Provider *string
	Version  *string
}

func (r *CodePipelineCustomActionType) Filter() error {
	if !strings.HasPrefix(*r.Owner, "Custom") {
		return fmt.Errorf("cannot delete default codepipeline custom action type")
	}
	return nil
}

func (r *CodePipelineCustomActionType) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCustomActionType(&codepipeline.DeleteCustomActionTypeInput{
		Category: r.Category,
		Provider: r.Provider,
		Version:  r.Version,
	})

	return err
}

func (r *CodePipelineCustomActionType) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodePipelineCustomActionType) String() string {
	return *r.Owner
}
