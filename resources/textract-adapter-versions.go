package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"                            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/textract"               //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/textract/textractiface" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TextractAdapterVersionResource = "TextractAdapterVersion"

func init() {
	registry.Register(&registry.Registration{
		Name:     TextractAdapterVersionResource,
		Scope:    nuke.Account,
		Resource: &TextractAdapterVersion{},
		Lister:   &TextractAdapterVersionLister{},
	})
}

type TextractAdapterVersionLister struct {
	mockSvc textractiface.TextractAPI
}

func (l *TextractAdapterVersionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc textractiface.TextractAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = textract.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	// First, list all adapters
	adapters, err := listTextractAdaptersForVersions(svc)
	if err != nil {
		return nil, err
	}

	// For each adapter, list its versions
	for _, adapter := range adapters {
		params := &textract.ListAdapterVersionsInput{
			AdapterId:  adapter.AdapterId,
			MaxResults: aws.Int64(100),
		}

		for {
			resp, err := svc.ListAdapterVersions(params)
			if err != nil {
				return nil, err
			}

			for _, item := range resp.AdapterVersions {
				resources = append(resources, &TextractAdapterVersion{
					svc:            svc,
					AdapterID:      item.AdapterId,
					AdapterVersion: item.AdapterVersion,
					Status:         item.Status,
					StatusMessage:  item.StatusMessage,
					CreationTime:   item.CreationTime,
					FeatureTypes:   item.FeatureTypes,
				})
			}

			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

// listTextractAdaptersForVersions lists all Textract adapters for the version lister
func listTextractAdaptersForVersions(svc textractiface.TextractAPI) ([]*textract.AdapterOverview, error) {
	adapters := make([]*textract.AdapterOverview, 0)
	params := &textract.ListAdaptersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListAdapters(params)
		if err != nil {
			return nil, err
		}
		adapters = append(adapters, resp.Adapters...)
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return adapters, nil
}

type TextractAdapterVersion struct {
	svc            textractiface.TextractAPI
	AdapterID      *string
	AdapterVersion *string
	Status         *string
	StatusMessage  *string
	CreationTime   *time.Time
	FeatureTypes   []*string
}

func (r *TextractAdapterVersion) Filter() error {
	if r.Status != nil {
		switch *r.Status {
		case textract.AdapterVersionStatusCreationInProgress:
			return fmt.Errorf("cannot delete adapter version in CREATION_IN_PROGRESS state")
		case textract.AdapterVersionStatusCreationError:
			return fmt.Errorf("cannot delete adapter version in CREATION_ERROR state")
		}
	}
	return nil
}

func (r *TextractAdapterVersion) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAdapterVersion(&textract.DeleteAdapterVersionInput{
		AdapterId:      r.AdapterID,
		AdapterVersion: r.AdapterVersion,
	})
	return err
}

func (r *TextractAdapterVersion) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *TextractAdapterVersion) String() string {
	return fmt.Sprintf("%s:%s", *r.AdapterID, *r.AdapterVersion)
}
