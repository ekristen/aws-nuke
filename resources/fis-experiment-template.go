package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FISExperimentTemplateResource = "FISExperimentTemplate"

func init() {
	registry.Register(&registry.Registration{
		Name:     FISExperimentTemplateResource,
		Scope:    nuke.Account,
		Resource: &FISExperimentTemplate{},
		Lister:   &FISExperimentTemplateLister{},
		DependsOn: []string{
			FISExperimentResource,
			FISTargetAccountConfigurationResource,
		},
		AlternativeResource: "AWS::FIS::ExperimentTemplate",
	})
}

type FISExperimentTemplateLister struct {
	svc FISAPI
}

func (l *FISExperimentTemplateLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = fis.NewFromConfig(*opts.Config)
	}

	var resources []resource.Resource

	params := &fis.ListExperimentTemplatesInput{
		MaxResults: aws.Int32(100),
	}

	for {
		resp, err := l.svc.ListExperimentTemplates(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.ExperimentTemplates {
			resources = append(resources, &FISExperimentTemplate{
				svc:            l.svc,
				ID:             item.Id,
				ARN:            item.Arn,
				Description:    item.Description,
				CreationTime:   item.CreationTime,
				LastUpdateTime: item.LastUpdateTime,
				Tags:           item.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type FISExperimentTemplate struct {
	svc FISAPI

	ID             *string           `description:"The ID of the experiment template"`
	ARN            *string           `description:"The ARN of the experiment template"`
	Description    *string           `description:"The description of the experiment template"`
	CreationTime   *time.Time        `description:"The time that the experiment template was created"`
	LastUpdateTime *time.Time        `description:"The time that the experiment template was last updated"`
	Tags           map[string]string `description:"The tags associated with the experiment template"`
}

func (r *FISExperimentTemplate) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteExperimentTemplate(ctx, &fis.DeleteExperimentTemplateInput{
		Id: r.ID,
	})
	return err
}

func (r *FISExperimentTemplate) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *FISExperimentTemplate) String() string {
	return *r.ID
}
