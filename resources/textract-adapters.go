package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	textracttypes "github.com/aws/aws-sdk-go-v2/service/textract/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TextractAdapterResource = "TextractAdapter"

func init() {
	registry.Register(&registry.Registration{
		Name:     TextractAdapterResource,
		Scope:    nuke.Account,
		Resource: &TextractAdapter{},
		Lister:   &TextractAdapterLister{},
	})
}

type TextractAdapterLister struct{}

func (l *TextractAdapterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := textract.NewFromConfig(*opts.Config)

	resources := make([]resource.Resource, 0)

	params := &textract.ListAdaptersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := textract.NewListAdaptersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, adapter := range resp.Adapters {
			// Get detailed adapter info including tags
			adapterDetails, err := svc.GetAdapter(ctx, &textract.GetAdapterInput{
				AdapterId: adapter.AdapterId,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &TextractAdapter{
				svc:          svc,
				AdapterID:    adapter.AdapterId,
				AdapterName:  adapter.AdapterName,
				CreationTime: adapter.CreationTime,
				FeatureTypes: adapter.FeatureTypes,
				AutoUpdate:   adapterDetails.AutoUpdate,
				Description:  adapterDetails.Description,
				Tags:         adapterDetails.Tags,
			})
		}
	}

	return resources, nil
}

type TextractAdapter struct {
	svc          *textract.Client
	AdapterID    *string
	AdapterName  *string
	CreationTime *time.Time
	FeatureTypes []textracttypes.FeatureType
	AutoUpdate   textracttypes.AutoUpdate
	Description  *string
	Tags         map[string]string
}

func (r *TextractAdapter) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAdapter(ctx, &textract.DeleteAdapterInput{
		AdapterId: r.AdapterID,
	})
	return err
}

func (r *TextractAdapter) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TextractAdapter) String() string {
	return *r.AdapterID
}
