package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/translate"
	translatetypes "github.com/aws/aws-sdk-go-v2/service/translate/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TranslateParallelDataResource = "TranslateParallelData"

func init() {
	registry.Register(&registry.Registration{
		Name:     TranslateParallelDataResource,
		Scope:    nuke.Account,
		Resource: &TranslateParallelData{},
		Lister:   &TranslateParallelDataLister{},
	})
}

type TranslateParallelDataLister struct {
	svc TranslateAPI
}

func (l *TranslateParallelDataLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		l.svc = translate.NewFromConfig(*opts.Config)
	}

	params := &translate.ListParallelDataInput{
		MaxResults: aws.Int32(500),
	}

	for {
		resp, err := l.svc.ListParallelData(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.ParallelDataPropertiesList {
			resources = append(resources, &TranslateParallelData{
				svc:                l.svc,
				Name:               item.Name,
				Arn:                item.Arn,
				Status:             item.Status,
				SourceLanguageCode: item.SourceLanguageCode,
				Description:        item.Description,
				CreatedAt:          item.CreatedAt,
				LastUpdatedAt:      item.LastUpdatedAt,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TranslateParallelData struct {
	svc                TranslateAPI
	Name               *string
	Arn                *string
	Status             translatetypes.ParallelDataStatus
	SourceLanguageCode *string
	Description        *string
	CreatedAt          *time.Time
	LastUpdatedAt      *time.Time
}

func (r *TranslateParallelData) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteParallelData(ctx, &translate.DeleteParallelDataInput{
		Name: r.Name,
	})
	return err
}

func (r *TranslateParallelData) Filter() error {
	if r.Status == translatetypes.ParallelDataStatusDeleting {
		return fmt.Errorf("parallel data is already deleting")
	}
	if r.Status == translatetypes.ParallelDataStatusFailed {
		return fmt.Errorf("parallel data is in failed state")
	}
	return nil
}

func (r *TranslateParallelData) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TranslateParallelData) String() string {
	return *r.Name
}
