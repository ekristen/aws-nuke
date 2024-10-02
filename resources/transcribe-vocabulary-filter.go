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

type TranscribeVocabularyFilter struct {
	svc              *transcribeservice.TranscribeService
	name             *string
	languageCode     *string
	lastModifiedTime *time.Time
}

func (r *TranscribeVocabularyFilter) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteVocabularyFilterInput{
		VocabularyFilterName: r.name,
	}
	_, err := r.svc.DeleteVocabularyFilter(deleteInput)
	return err
}

func (r *TranscribeVocabularyFilter) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("LanguageCode", r.languageCode)
	properties.Set("LastModifiedTime", r.lastModifiedTime)
	return properties
}
