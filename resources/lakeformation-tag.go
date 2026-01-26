package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LakeFormationTagResource = "LakeFormationTag"

func init() {
	registry.Register(&registry.Registration{
		Name:     LakeFormationTagResource,
		Scope:    nuke.Account,
		Resource: &LakeFormationTag{},
		Lister:   &LakeFormationTagLister{},
	})
}

type LakeFormationTagLister struct{}

func (l *LakeFormationTagLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lakeformation.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := lakeformation.NewListLFTagsPaginator(svc, &lakeformation.ListLFTagsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, t := range page.LFTags {
			resources = append(resources, &LakeFormationTag{
				svc:       svc,
				TagKey:    t.TagKey,
				CatalogID: t.CatalogId,
			})
		}
	}

	return resources, nil
}

type LakeFormationTag struct {
	svc       *lakeformation.Client
	TagKey    *string `description:"The key-name for the LF-tag"`
	CatalogID *string `description:"The identifier for the Data Catalog. By default, the account ID."`
}

func (f *LakeFormationTag) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteLFTag(ctx, &lakeformation.DeleteLFTagInput{
		TagKey:    f.TagKey,
		CatalogId: f.CatalogID,
	})

	return err
}

func (f *LakeFormationTag) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *LakeFormationTag) String() string {
	return *f.TagKey
}
