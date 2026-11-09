package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppConfigExtensionResource = "AppConfigExtension"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppConfigExtensionResource,
		Scope:    nuke.Account,
		Resource: &AppConfigExtension{},
		Lister:   &AppConfigExtensionLister{},
		DependsOn: []string{
			AppConfigExtensionAssociationResource,
		},
	})
}

type AppConfigExtensionLister struct {
	svc AppConfigAPI
}

func (l *AppConfigExtensionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)

	svc := l.svc
	if svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc = appconfig.NewFromConfig(*opts.Config)
	}

	paginator := appconfig.NewListExtensionsPaginator(svc, &appconfig.ListExtensionsInput{
		MaxResults: aws.Int32(50),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range page.Items {
			item := &page.Items[i]
			resources = append(resources, &AppConfigExtension{
				svc:           svc,
				ID:            item.Id,
				Name:          item.Name,
				ARN:           item.Arn,
				VersionNumber: aws.Int32(item.VersionNumber),
			})
		}
	}

	return resources, nil
}

type AppConfigExtension struct {
	svc           AppConfigAPI
	ID            *string `description:"The ID of the extension"`
	Name          *string `description:"The name of the extension"`
	ARN           *string `description:"The ARN of the extension"`
	VersionNumber *int32  `description:"The version number of the extension"`
}

// Filter skips the extensions authored by AWS (for example AWS.AppConfig.FeatureFlags). They are returned by
// ListExtensions but cannot be deleted.
func (r *AppConfigExtension) Filter() error {
	if strings.HasPrefix(aws.ToString(r.Name), "AWS.") {
		return fmt.Errorf("cannot delete AWS authored extension")
	}
	return nil
}

func (r *AppConfigExtension) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteExtension(ctx, &appconfig.DeleteExtensionInput{
		ExtensionIdentifier: r.ID,
		VersionNumber:       r.VersionNumber,
	})
	return err
}

func (r *AppConfigExtension) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppConfigExtension) String() string {
	return aws.ToString(r.Name)
}
