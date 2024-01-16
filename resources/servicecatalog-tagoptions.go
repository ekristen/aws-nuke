package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogTagOptionResource = "ServiceCatalogTagOption"

func init() {
	resource.Register(resource.Registration{
		Name:   ServiceCatalogTagOptionResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogTagOptionLister{},
	})
}

type ServiceCatalogTagOptionLister struct{}

func (l *ServiceCatalogTagOptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &servicecatalog.ListTagOptionsInput{
		PageSize: aws.Int64(20),
	}

	for {
		resp, err := svc.ListTagOptions(params)
		if err != nil {
			if awsutil.IsAWSError(err, servicecatalog.ErrCodeTagOptionNotMigratedException) {
				logrus.Info(err)
				break
			}
			return nil, err
		}

		for _, tagOptionDetail := range resp.TagOptionDetails {
			resources = append(resources, &ServiceCatalogTagOption{
				svc:   svc,
				ID:    tagOptionDetail.Id,
				key:   tagOptionDetail.Key,
				value: tagOptionDetail.Value,
			})
		}

		if resp.PageToken == nil {
			break
		}

		params.PageToken = resp.PageToken
	}

	return resources, nil
}

type ServiceCatalogTagOption struct {
	svc   *servicecatalog.ServiceCatalog
	ID    *string
	key   *string
	value *string
}

func (f *ServiceCatalogTagOption) Remove(_ context.Context) error {
	_, err := f.svc.DeleteTagOption(&servicecatalog.DeleteTagOptionInput{
		Id: f.ID,
	})

	return err
}

func (f *ServiceCatalogTagOption) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Key", f.key)
	properties.Set("Value", f.value)
	return properties
}

func (f *ServiceCatalogTagOption) String() string {
	return *f.ID
}
