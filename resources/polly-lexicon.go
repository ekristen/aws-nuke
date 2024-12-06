package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/polly"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const PollyLexiconResource = "PollyLexicon"

func init() {
	registry.Register(&registry.Registration{
		Name:     PollyLexiconResource,
		Scope:    nuke.Account,
		Resource: &PollyLexicon{},
		Lister:   &PollyLexiconLister{},
	})
}

type PollyLexiconLister struct{}

func (l *PollyLexiconLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := polly.New(opts.Session)

	params := &polly.ListLexiconsInput{}

	for {
		listOutput, err := svc.ListLexicons(params)
		if err != nil {
			return nil, err
		}
		for _, lexicon := range listOutput.Lexicons {
			resources = append(resources, &PollyLexicon{
				svc:          svc,
				Name:         lexicon.Name,
				Alphabet:     lexicon.Attributes.Alphabet,
				LanguageCode: lexicon.Attributes.LanguageCode,
				LastModified: lexicon.Attributes.LastModified,
				LexemesCount: lexicon.Attributes.LexemesCount,
				LexiconArn:   lexicon.Attributes.LexiconArn,
				Size:         lexicon.Attributes.Size,
			})
		}

		// Check if there are more results
		if listOutput.NextToken == nil {
			break // No more results, exit the loop
		}

		// Set the nextToken for the next iteration
		params.NextToken = listOutput.NextToken
	}

	return resources, nil
}

type PollyLexicon struct {
	svc          *polly.Polly
	Name         *string
	Alphabet     *string
	LanguageCode *string
	LastModified *time.Time
	LexemesCount *int64
	LexiconArn   *string
	Size         *int64
}

func (r *PollyLexicon) Remove(_ context.Context) error {
	_, err := r.svc.DeleteLexicon(&polly.DeleteLexiconInput{
		Name: r.Name,
	})
	return err
}

func (r *PollyLexicon) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PollyLexicon) String() string {
	return *r.Name
}
