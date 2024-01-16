package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogProductResource = "ServiceCatalogProduct"

func init() {
	resource.Register(resource.Registration{
		Name:   ServiceCatalogProductResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogProductLister{},
	})
}

type ServiceCatalogProductLister struct{}

func (l *ServiceCatalogProductLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &servicecatalog.SearchProductsAsAdminInput{
		PageSize: aws.Int64(20),
	}

	for {
		resp, err := svc.SearchProductsAsAdmin(params)
		if err != nil {
			return nil, err
		}

		for _, productView := range resp.ProductViewDetails {
			resources = append(resources, &ServiceCatalogProduct{
				svc:  svc,
				ID:   productView.ProductViewSummary.ProductId,
				name: productView.ProductViewSummary.Name,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ServiceCatalogProduct struct {
	svc  *servicecatalog.ServiceCatalog
	ID   *string
	name *string
}

func (f *ServiceCatalogProduct) Remove(_ context.Context) error {
	_, err := f.svc.DeleteProduct(&servicecatalog.DeleteProductInput{
		Id: f.ID,
	})

	return err
}

func (f *ServiceCatalogProduct) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Name", f.name)
	return properties
}

func (f *ServiceCatalogProduct) String() string {
	return *f.ID
}
