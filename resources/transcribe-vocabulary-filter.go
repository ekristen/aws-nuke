package resources

import (
	"context"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	
	"github.com/aws/aws-sdk-go/service/transcribeservice"
	""
)

type TranscribeVocabularyFilter struct {
	svc              *transcribeservice.TranscribeService
	name             *string
	languageCode     *string
	lastModifiedTime *time.Time
}

const TranscribeVocabularyFilterResource = "TranscribeVocabularyFilter"

func init() {
	registry.Register(&registry.Registration{
		Name:   TranscribeVocabularyFilterResource,
		Scope:  nuke.Account,
		Lister: &TranscribeVocabularyFilterLister{},
	})
}

type TranscribeVocabularyFilterLister struct{}

func (l *TranscribeVocabularyFilterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listVocabularyFiltersInput := &transcribeservice.ListVocabularyFiltersInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}

		listOutput, err := svc.ListVocabularyFilters(listVocabularyFiltersInput)
		if err != nil {
			return nil, err
		}
		for _, filter := range listOutput.VocabularyFilters {
			resources = append(resources, &TranscribeVocabularyFilter{
				svc:              svc,
				name:             filter.VocabularyFilterName,
				languageCode:     filter.LanguageCode,
				lastModifiedTime: filter.LastModifiedTime,
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

func (filter *TranscribeVocabularyFilter) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteVocabularyFilterInput{
		VocabularyFilterName: filter.name,
	}
	_, err := filter.svc.DeleteVocabularyFilter(deleteInput)
	return err
}

func (filter *TranscribeVocabularyFilter) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", filter.name)
	properties.Set("LanguageCode", filter.languageCode)
	if filter.lastModifiedTime != nil {
		properties.Set("LastModifiedTime", filter.lastModifiedTime.Format(time.RFC3339))
	}
	return properties
}
