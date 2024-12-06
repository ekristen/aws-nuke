package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotsitewise"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTSiteWiseAssetModelResource = "IoTSiteWiseAssetModel"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTSiteWiseAssetModelResource,
		Scope:    nuke.Account,
		Resource: &IoTSiteWiseAssetModel{},
		Lister:   &IoTSiteWiseAssetModelLister{},
		DependsOn: []string{
			IoTSiteWiseAssetResource,
		},
	})
}

type IoTSiteWiseAssetModelLister struct{}

func (l *IoTSiteWiseAssetModelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iotsitewise.ListAssetModelsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListAssetModels(params)
		if err != nil {
			return nil, err
		}
		for _, item := range resp.AssetModelSummaries {
			tagResp, err := svc.ListTagsForResource(
				&iotsitewise.ListTagsForResourceInput{
					ResourceArn: item.Arn,
				})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &IoTSiteWiseAssetModel{
				svc:    svc,
				ID:     item.Id,
				Name:   item.Name,
				Status: item.Status.State,
				Tags:   tagResp.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type IoTSiteWiseAssetModel struct {
	svc    *iotsitewise.IoTSiteWise
	ID     *string
	Name   *string
	Status *string
	Tags   map[string]*string
}

func (r *IoTSiteWiseAssetModel) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWiseAssetModel) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAssetModel(&iotsitewise.DeleteAssetModelInput{
		AssetModelId: r.ID,
	})
	return err
}

func (r *IoTSiteWiseAssetModel) String() string {
	return *r.ID
}
