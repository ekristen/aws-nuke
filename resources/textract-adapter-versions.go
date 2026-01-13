package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	textracttypes "github.com/aws/aws-sdk-go-v2/service/textract/types"

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

type TextractAdapterVersionLister struct{}

func (l *TextractAdapterVersionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := textract.NewFromConfig(*opts.Config)

	resources := make([]resource.Resource, 0)

	// First, list all adapters
	adapterParams := &textract.ListAdaptersInput{
		MaxResults: aws.Int32(100),
	}

	adapterPaginator := textract.NewListAdaptersPaginator(svc, adapterParams)

	for adapterPaginator.HasMorePages() {
		adapterResp, err := adapterPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// For each adapter, list its versions
		for _, adapter := range adapterResp.Adapters {
			versionParams := &textract.ListAdapterVersionsInput{
				AdapterId:  adapter.AdapterId,
				MaxResults: aws.Int32(100),
			}

			versionPaginator := textract.NewListAdapterVersionsPaginator(svc, versionParams)

			for versionPaginator.HasMorePages() {
				versionResp, err := versionPaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, item := range versionResp.AdapterVersions {
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
			}
		}
	}

	return resources, nil
}

type TextractAdapterVersion struct {
	svc            *textract.Client
	AdapterID      *string
	AdapterVersion *string
	Status         textracttypes.AdapterVersionStatus
	StatusMessage  *string
	CreationTime   *time.Time
	FeatureTypes   []textracttypes.FeatureType
}

func (r *TextractAdapterVersion) Filter() error {
	switch r.Status {
	case textracttypes.AdapterVersionStatusCreationInProgress:
		return fmt.Errorf("cannot delete adapter version in CREATION_IN_PROGRESS state")
	case textracttypes.AdapterVersionStatusCreationError:
		return fmt.Errorf("cannot delete adapter version in CREATION_ERROR state")
	}
	return nil
}

func (r *TextractAdapterVersion) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAdapterVersion(ctx, &textract.DeleteAdapterVersionInput{
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
