package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"                            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/textract"               //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/textract/textractiface" //nolint:staticcheck

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

type TextractAdapterLister struct {
	mockSvc textractiface.TextractAPI
}

func (l *TextractAdapterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc textractiface.TextractAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = textract.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	params := &textract.ListAdaptersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListAdapters(params)
		if err != nil {
			return nil, err
		}

		for _, adapter := range resp.Adapters {
			// Get detailed adapter info including tags
			adapterDetails, err := svc.GetAdapter(&textract.GetAdapterInput{
				AdapterId: adapter.AdapterId,
			})
			if err != nil {
				return nil, err
			}

			// Convert tags from map[string]*string to map[string]string
			tags := make(map[string]string)
			for k, v := range adapterDetails.Tags {
				if v != nil {
					tags[k] = *v
				}
			}

			resources = append(resources, &TextractAdapter{
				svc:          svc,
				AdapterID:    adapter.AdapterId,
				AdapterName:  adapter.AdapterName,
				CreationTime: adapter.CreationTime,
				FeatureTypes: adapter.FeatureTypes,
				AutoUpdate:   adapterDetails.AutoUpdate,
				Description:  adapterDetails.Description,
				Tags:         tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type TextractAdapter struct {
	svc          textractiface.TextractAPI
	AdapterID    *string
	AdapterName  *string
	CreationTime *time.Time
	FeatureTypes []*string
	AutoUpdate   *string
	Description  *string
	Tags         map[string]string
}

func (r *TextractAdapter) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAdapter(&textract.DeleteAdapterInput{
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
