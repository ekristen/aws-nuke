package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LakeFormationLocationResource = "LakeFormationLocation"

func init() {
	registry.Register(&registry.Registration{
		Name:     LakeFormationLocationResource,
		Scope:    nuke.Account,
		Resource: &LakeFormationLocation{},
		Lister:   &LakeFormationLocationLister{},
	})
}

type LakeFormationLocationLister struct{}

func (l *LakeFormationLocationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lakeformation.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := lakeformation.NewListResourcesPaginator(svc, &lakeformation.ListResourcesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, ri := range page.ResourceInfoList {
			resources = append(resources, &LakeFormationLocation{
				svc:         svc,
				ResourceARN: ri.ResourceArn,
			})
		}
	}

	return resources, nil
}

type LakeFormationLocation struct {
	svc         *lakeformation.Client
	ResourceARN *string `description:"The ARN of the resource registered with Lake Formation"`
}

func (f *LakeFormationLocation) Remove(ctx context.Context) error {
	_, err := f.svc.DeregisterResource(ctx, &lakeformation.DeregisterResourceInput{
		ResourceArn: f.ResourceARN,
	})

	return err
}

func (f *LakeFormationLocation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}
