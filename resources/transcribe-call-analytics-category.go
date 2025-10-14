package resources

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/transcribeservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranscribeCallAnalyticsCategoryResource = "TranscribeCallAnalyticsCategory"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranscribeCallAnalyticsCategoryResource,
		Scope:    nuke.Account,
		Resource: &TranscribeCallAnalyticsCategory{},
		Lister:   &TranscribeCallAnalyticsCategoryLister{},
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
			var badRequestException *transcribeservice.BadRequestException
			if errors.As(err, &badRequestException) &&
				strings.Contains(badRequestException.Message(), "isn't supported in this region") {
				return resources, nil
			}
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
	properties.Set("CreateTime", r.createTime)
	properties.Set("LastUpdateTime", r.lastUpdateTime)
	return properties
}

func (r *TranscribeCallAnalyticsCategory) String() string {
	return *r.name
}
