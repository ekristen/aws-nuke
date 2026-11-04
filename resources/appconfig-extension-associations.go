package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppConfigExtensionAssociationResource = "AppConfigExtensionAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppConfigExtensionAssociationResource,
		Scope:    nuke.Account,
		Resource: &AppConfigExtensionAssociation{},
		Lister:   &AppConfigExtensionAssociationLister{},
	})
}

type AppConfigExtensionAssociationLister struct {
	svc AppConfigAPI
}

func (l *AppConfigExtensionAssociationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)

	svc := l.svc
	if svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc = appconfig.NewFromConfig(*opts.Config)
	}

	paginator := appconfig.NewListExtensionAssociationsPaginator(svc, &appconfig.ListExtensionAssociationsInput{
		MaxResults: aws.Int32(50),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range page.Items {
			item := &page.Items[i]
			resources = append(resources, &AppConfigExtensionAssociation{
				svc:          svc,
				ID:           item.Id,
				ExtensionARN: item.ExtensionArn,
				ResourceARN:  item.ResourceArn,
			})
		}
	}

	return resources, nil
}

type AppConfigExtensionAssociation struct {
	svc          AppConfigAPI
	ID           *string `description:"The ID of the extension association"`
	ExtensionARN *string `description:"The ARN of the associated extension"`
	ResourceARN  *string `description:"The ARN of the application, configuration profile or environment the extension is associated with"`
}

func (r *AppConfigExtensionAssociation) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteExtensionAssociation(ctx, &appconfig.DeleteExtensionAssociationInput{
		ExtensionAssociationId: r.ID,
	})
	return err
}

func (r *AppConfigExtensionAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppConfigExtensionAssociation) String() string {
	return aws.ToString(r.ID)
}
