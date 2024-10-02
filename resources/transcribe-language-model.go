package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transcribeservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranscribeLanguageModelResource = "TranscribeLanguageModel"

func init() {
	registry.Register(&registry.Registration{
		Name:   TranscribeLanguageModelResource,
		Scope:  nuke.Account,
		Lister: &TranscribeLanguageModelLister{},
	})
}

type TranscribeLanguageModelLister struct{}

func (l *TranscribeLanguageModelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listLanguageModelsInput := &transcribeservice.ListLanguageModelsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}

		listOutput, err := svc.ListLanguageModels(listLanguageModelsInput)
		if err != nil {
			return nil, err
		}
		for _, model := range listOutput.Models {
			resources = append(resources, &TranscribeLanguageModel{
				svc:                 svc,
				name:                model.ModelName,
				baseModelName:       model.BaseModelName,
				createTime:          model.CreateTime,
				failureReason:       model.FailureReason,
				languageCode:        model.LanguageCode,
				lastModifiedTime:    model.LastModifiedTime,
				modelStatus:         model.ModelStatus,
				upgradeAvailability: model.UpgradeAvailability,
			})
		}

		// Check if there are more results
		if listOutput.NextToken == nil {
			break // No more results, exit the loop
		}

		// Set the nextToken for the next iteration
		nextToken = listOutput.NextToken
	}
	return resources, nil
}

type TranscribeLanguageModel struct {
	svc                 *transcribeservice.TranscribeService
	name                *string
	baseModelName       *string
	createTime          *time.Time
	failureReason       *string
	languageCode        *string
	lastModifiedTime    *time.Time
	modelStatus         *string
	upgradeAvailability *bool
}

func (r *TranscribeLanguageModel) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteLanguageModelInput{
		ModelName: r.name,
	}
	_, err := r.svc.DeleteLanguageModel(deleteInput)
	return err
}

func (r *TranscribeLanguageModel) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("BaseModelName", r.baseModelName)
	properties.Set("CreateTime", r.createTime)
	properties.Set("FailureReason", r.failureReason)
	properties.Set("LanguageCode", r.languageCode)
	properties.Set("LastModifiedTime", r.lastModifiedTime)
	properties.Set("ModelStatus", r.modelStatus)
	properties.Set("UpgradeAvailability", r.upgradeAvailability)
	return properties
}

func (r *TranscribeLanguageModel) String() string {
	return *r.name
}
