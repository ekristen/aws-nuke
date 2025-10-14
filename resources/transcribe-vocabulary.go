package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"                       //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/transcribeservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranscribeVocabularyResource = "TranscribeVocabulary"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranscribeVocabularyResource,
		Scope:    nuke.Account,
		Resource: &TranscribeVocabulary{},
		Lister:   &TranscribeVocabularyLister{},
	})
}

type TranscribeVocabularyLister struct{}

func (l *TranscribeVocabularyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transcribeservice.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		listVocabulariesInput := &transcribeservice.ListVocabulariesInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}

		listOutput, err := svc.ListVocabularies(listVocabulariesInput)
		if err != nil {
			return nil, err
		}
		for _, vocab := range listOutput.Vocabularies {
			resources = append(resources, &TranscribeVocabulary{
				svc:              svc,
				name:             vocab.VocabularyName,
				state:            vocab.VocabularyState,
				languageCode:     vocab.LanguageCode,
				lastModifiedTime: vocab.LastModifiedTime,
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

type TranscribeVocabulary struct {
	svc              *transcribeservice.TranscribeService
	name             *string
	state            *string
	languageCode     *string
	lastModifiedTime *time.Time
}

func (r *TranscribeVocabulary) Remove(_ context.Context) error {
	deleteInput := &transcribeservice.DeleteVocabularyInput{
		VocabularyName: r.name,
	}
	_, err := r.svc.DeleteVocabulary(deleteInput)
	return err
}

func (r *TranscribeVocabulary) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("State", r.state)
	properties.Set("LanguageCode", r.languageCode)
	properties.Set("LastModifiedTime", r.lastModifiedTime)
	return properties
}

func (r *TranscribeVocabulary) String() string {
	return *r.name
}
