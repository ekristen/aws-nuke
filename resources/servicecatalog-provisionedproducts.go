package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogProvisionedProductResource = "ServiceCatalogProvisionedProduct"

func init() {
	resource.Register(resource.Registration{
		Name:   ServiceCatalogProvisionedProductResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogProvisionedProductLister{},
	})
}

type ServiceCatalogProvisionedProductLister struct{}

func (l *ServiceCatalogProvisionedProductLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &servicecatalog.ScanProvisionedProductsInput{
		PageSize: aws.Int64(20),
		AccessLevelFilter: &servicecatalog.AccessLevelFilter{
			Key:   aws.String("Account"),
			Value: aws.String("self"),
		},
	}

	for {
		resp, err := svc.ScanProvisionedProducts(params)
		if err != nil {
			return nil, err
		}

		for _, provisionedProduct := range resp.ProvisionedProducts {
			resources = append(resources, &ServiceCatalogProvisionedProduct{
				svc:            svc,
				ID:             provisionedProduct.Id,
				terminateToken: provisionedProduct.IdempotencyToken,
				name:           provisionedProduct.Name,
				productID:      provisionedProduct.ProductId,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ServiceCatalogProvisionedProduct struct {
	svc            *servicecatalog.ServiceCatalog
	ID             *string
	terminateToken *string
	name           *string
	productID      *string
}

func (f *ServiceCatalogProvisionedProduct) Remove(_ context.Context) error {
	_, err := f.svc.TerminateProvisionedProduct(&servicecatalog.TerminateProvisionedProductInput{
		ProvisionedProductId: f.ID,
		IgnoreErrors:         aws.Bool(true),
		TerminateToken:       f.terminateToken,
	})

	return err
}

func (f *ServiceCatalogProvisionedProduct) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Name", f.name)
	properties.Set("ProductID", f.productID)
	return properties
}

func (f *ServiceCatalogProvisionedProduct) String() string {
	return *f.ID
}
