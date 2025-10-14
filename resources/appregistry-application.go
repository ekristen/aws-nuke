package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/appregistry" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppRegistryApplicationResource = "AppRegistryApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppRegistryApplicationResource,
		Scope:    nuke.Account,
		Resource: &AppRegistryApplication{},
		Lister:   &AppRegistryApplicationLister{},
	})
}

type AppRegistryApplicationLister struct{}

func (l *AppRegistryApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := appregistry.New(opts.Session)
	var resources []resource.Resource

	res, err := svc.ListApplications(&appregistry.ListApplicationsInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.Applications {
		tags, err := svc.ListTagsForResource(&appregistry.ListTagsForResourceInput{
			ResourceArn: p.Arn,
		})
		if err != nil {
			logrus.WithError(err).Error("unable to get tags for AppRegistry Application")
		}

		newResource := &AppRegistryApplication{
			svc: svc,
			ID:  p.Id,
		}

		if tags != nil {
			for key, val := range tags.Tags {
				if key == "aws:servicecatalog:applicationName" {
					newResource.Name = val
					break
				}
			}

			newResource.Tags = tags.Tags
		}

		resources = append(resources, newResource)
	}

	return resources, nil
}

type AppRegistryApplication struct {
	svc  *appregistry.AppRegistry
	ID   *string
	Name *string
	Tags map[string]*string
}

func (r *AppRegistryApplication) Remove(_ context.Context) error {
	_, err := r.svc.DeleteApplication(&appregistry.DeleteApplicationInput{
		Application: r.ID,
	})
	return err
}

func (r *AppRegistryApplication) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AppRegistryApplication) String() string {
	return *r.Name
}
