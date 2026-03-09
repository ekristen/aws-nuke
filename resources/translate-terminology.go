package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/translate"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranslateTerminologyResource = "TranslateTerminology"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranslateTerminologyResource,
		Scope:    nuke.Account,
		Resource: &TranslateTerminology{},
		Lister:   &TranslateTerminologyLister{},
	})
}

type TranslateTerminologyLister struct {
	svc TranslateAPI
}

func (l *TranslateTerminologyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = translate.NewFromConfig(*opts.Config)
	}

	params := &translate.ListTerminologiesInput{
		MaxResults: aws.Int32(500),
	}

	for {
		resp, err := l.svc.ListTerminologies(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.TerminologyPropertiesList {
			item := &resp.TerminologyPropertiesList[i]
			resources = append(resources, &TranslateTerminology{
				svc:                l.svc,
				Name:               item.Name,
				Arn:                item.Arn,
				SourceLanguageCode: item.SourceLanguageCode,
				Description:        item.Description,
				CreatedAt:          item.CreatedAt,
				LastUpdatedAt:      item.LastUpdatedAt,
				SizeBytes:          item.SizeBytes,
				TermCount:          item.TermCount,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TranslateTerminology struct {
	svc                TranslateAPI
	Name               *string
	Arn                *string
	SourceLanguageCode *string
	Description        *string
	CreatedAt          *time.Time
	LastUpdatedAt      *time.Time
	SizeBytes          *int32
	TermCount          *int32
}

func (r *TranslateTerminology) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteTerminology(ctx, &translate.DeleteTerminologyInput{
		Name: r.Name,
	})
	return err
}

func (r *TranslateTerminology) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TranslateTerminology) String() string {
	return *r.Name
}
