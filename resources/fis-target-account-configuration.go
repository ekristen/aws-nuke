package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	fistypes "github.com/aws/aws-sdk-go-v2/service/fis/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FISTargetAccountConfigurationResource = "FISTargetAccountConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:                FISTargetAccountConfigurationResource,
		Scope:               nuke.Account,
		Resource:            &FISTargetAccountConfiguration{},
		Lister:              &FISTargetAccountConfigurationLister{},
		AlternativeResource: "AWS::FIS::TargetAccountConfiguration",
	})
}

type FISTargetAccountConfigurationLister struct {
	svc FISAPI
}

func (l *FISTargetAccountConfigurationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	svc := l.svc
	if svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc = fis.NewFromConfig(*opts.Config)
	}

	var resources []resource.Resource

	templateParams := &fis.ListExperimentTemplatesInput{
		MaxResults: aws.Int32(100),
	}

	for {
		templates, err := svc.ListExperimentTemplates(ctx, templateParams)
		if err != nil {
			return nil, err
		}

		for _, template := range templates.ExperimentTemplates {
			configParams := &fis.ListTargetAccountConfigurationsInput{
				ExperimentTemplateId: template.Id,
				MaxResults:           aws.Int32(100),
			}

			for {
				configs, err := svc.ListTargetAccountConfigurations(ctx, configParams)
				if err != nil {
					// Templates are deleted in the same run, so this template may already be gone.
					var notFound *fistypes.ResourceNotFoundException
					if errors.As(err, &notFound) {
						break
					}
					return nil, err
				}

				for _, item := range configs.TargetAccountConfigurations {
					resources = append(resources, &FISTargetAccountConfiguration{
						svc:                  svc,
						ExperimentTemplateID: template.Id,
						AccountID:            item.AccountId,
						RoleARN:              item.RoleArn,
						Description:          item.Description,
					})
				}

				if configs.NextToken == nil {
					break
				}

				configParams.NextToken = configs.NextToken
			}
		}

		if templates.NextToken == nil {
			break
		}

		templateParams.NextToken = templates.NextToken
	}

	return resources, nil
}

type FISTargetAccountConfiguration struct {
	svc FISAPI

	ExperimentTemplateID *string `description:"The ID of the experiment template"`
	AccountID            *string `description:"The AWS account ID of the target account"`
	RoleARN              *string `description:"The ARN of the IAM role for the target account"`
	Description          *string `description:"The description of the target account configuration"`
}

func (r *FISTargetAccountConfiguration) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteTargetAccountConfiguration(ctx, &fis.DeleteTargetAccountConfigurationInput{
		ExperimentTemplateId: r.ExperimentTemplateID,
		AccountId:            r.AccountID,
	})
	return err
}

func (r *FISTargetAccountConfiguration) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *FISTargetAccountConfiguration) String() string {
	return fmt.Sprintf("%s:%s", *r.ExperimentTemplateID, *r.AccountID)
}
