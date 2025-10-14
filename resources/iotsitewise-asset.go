package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iotsitewise" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTSiteWiseAssetResource = "IoTSiteWiseAsset"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTSiteWiseAssetResource,
		Scope:    nuke.Account,
		Resource: &IoTSiteWiseAsset{},
		Lister:   &IoTSiteWiseAssetLister{},
	})
}

type IoTSiteWiseAssetLister struct{}

func (l *IoTSiteWiseAssetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	assetModelSummaries, err := ListAssetModels(svc)
	if err != nil {
		return nil, err
	}

	for _, assetModelSummary := range assetModelSummaries {
		params := &iotsitewise.ListAssetsInput{
			AssetModelId: assetModelSummary.Id,
			MaxResults:   aws.Int64(25),
		}

		for {
			resp, err := svc.ListAssets(params)
			if err != nil {
				return nil, err
			}
			for _, item := range resp.AssetSummaries {
				tagResp, err := svc.ListTagsForResource(
					&iotsitewise.ListTagsForResourceInput{
						ResourceArn: item.Arn,
					})
				if err != nil {
					return nil, err
				}

				resources = append(resources, &IoTSiteWiseAsset{
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
	}

	return resources, nil
}

// Utility function to get models, allowing to scan for assets
func ListAssetModels(svc *iotsitewise.IoTSiteWise) ([]*iotsitewise.AssetModelSummary, error) {
	resources := make([]*iotsitewise.AssetModelSummary, 0)
	params := &iotsitewise.ListAssetModelsInput{
		MaxResults: aws.Int64(25),
	}
	for {
		resp, err := svc.ListAssetModels(params)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resp.AssetModelSummaries...)
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type IoTSiteWiseAsset struct {
	svc    *iotsitewise.IoTSiteWise
	ID     *string
	Name   *string
	Status *string
	Tags   map[string]*string
}

func (r *IoTSiteWiseAsset) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWiseAsset) Remove(_ context.Context) error {
	associatedAssets, err := r.svc.ListAssociatedAssets(&iotsitewise.ListAssociatedAssetsInput{
		AssetId: r.ID,
	})
	if err != nil {
		return err
	}
	assetDescription, err := r.svc.DescribeAsset(&iotsitewise.DescribeAssetInput{
		AssetId: r.ID,
	})
	if err != nil {
		return err
	}

	// If asset is associated, dissociate before delete
	for _, assetHierarchy := range assetDescription.AssetHierarchies {
		for _, childAsset := range associatedAssets.AssetSummaries {
			// Could fail if hierarchy it not the correct one, ignore it
			_, _ = r.svc.DisassociateAssets(&iotsitewise.DisassociateAssetsInput{
				AssetId:      r.ID,
				ChildAssetId: childAsset.Id,
				HierarchyId:  assetHierarchy.Id,
			})
		}
	}

	_, err = r.svc.DeleteAsset(&iotsitewise.DeleteAssetInput{
		AssetId: r.ID,
	})

	return err
}

func (r *IoTSiteWiseAsset) String() string {
	return *r.ID
}
