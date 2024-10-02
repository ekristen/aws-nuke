package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/transcribeservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranscribeCallAnalyticsCategoryResource = "TranscribeCallAnalyticsCategory"

func init() {
	registry.Register(&registry.Registration{
		Name:   TranscribeCallAnalyticsCategoryResource,
		Scope:  nuke.Account,
		Lister: &TranscribeCallAnalyticsCategoryLister{},
	})
}

type TranscribeCallAnalyticsCategoryLister struct{}

func (l *TranscribeCallAnalyticsCategoryLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listCallAnalyticsCategoriesInput := &transcribeservice.ListCallAnalyticsCategoriesInput{
			NextToken: nextToken,
		}

		listOutput, err := svc.ListCallAnalyticsCategories(listCallAnalyticsCategoriesInput)
		if err != nil {
			return nil, err
		}
		for _, category := range listOutput.Categories {
			resources = append(resources, &TranscribeCallAnalyticsCategory{
				svc:            svc,
				name:           category.CategoryName,
				inputType:      category.InputType,
				createTime:     category.CreateTime,
				lastUpdateTime: category.LastUpdateTime,
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

type TranscribeCallAnalyticsCategory struct {
	svc            *transcribeservice.TranscribeService
	name           *string
	inputType      *string
	createTime     *time.Time
	lastUpdateTime *time.Time
}

func (r *TranscribeCallAnalyticsCategory) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteCallAnalyticsCategoryInput{
		CategoryName: r.name,
	}
	_, err := r.svc.DeleteCallAnalyticsCategory(deleteInput)
	return err
}

func (r *TranscribeCallAnalyticsCategory) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("InputType", r.inputType)
	if r.createTime != nil {
		properties.Set("CreateTime", r.createTime.Format(time.RFC3339))
	}
	if r.lastUpdateTime != nil {
		properties.Set("LastUpdateTime", r.lastUpdateTime.Format(time.RFC3339))
	}
	return properties
}

func (r *TranscribeCallAnalyticsCategory) String() string {
	return *r.name
}
