package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogPortfolioProductAttachmentResource = "ServiceCatalogPortfolioProductAttachment"

func init() {
	resource.Register(resource.Registration{
		Name:   ServiceCatalogPortfolioProductAttachmentResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogPortfolioProductAttachmentLister{},
	})
}

type ServiceCatalogPortfolioProductAttachmentLister struct{}

func (l *ServiceCatalogPortfolioProductAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)
	products := make(map[*string]*string)

	params := &servicecatalog.SearchProductsAsAdminInput{
		PageSize: aws.Int64(20),
	}

	// List all Products and then search assigned portfolios
	for {
		resp, err := svc.SearchProductsAsAdmin(params)
		if err != nil {
			return nil, err
		}

		for _, productViewDetail := range resp.ProductViewDetails {
			products[productViewDetail.ProductViewSummary.ProductId] = productViewDetail.ProductViewSummary.Name
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	portfolioParams := &servicecatalog.ListPortfoliosForProductInput{
		PageSize: aws.Int64(20),
	}

	for productID, productName := range products {

		portfolioParams.ProductId = productID

		resp, err := svc.ListPortfoliosForProduct(portfolioParams)
		if err != nil {
			return nil, err
		}

		for _, portfolioDetail := range resp.PortfolioDetails {
			resources = append(resources, &ServiceCatalogPortfolioProductAttachment{
				svc:           svc,
				productID:     productID,
				portfolioID:   portfolioDetail.Id,
				portfolioName: portfolioDetail.DisplayName,
				productName:   productName,
			})
		}

		if resp.NextPageToken == nil {
			continue
		}

		portfolioParams.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ServiceCatalogPortfolioProductAttachment struct {
	svc           *servicecatalog.ServiceCatalog
	productID     *string
	portfolioID   *string
	portfolioName *string
	productName   *string
}

func (f *ServiceCatalogPortfolioProductAttachment) Remove(_ context.Context) error {
	_, err := f.svc.DisassociateProductFromPortfolio(&servicecatalog.DisassociateProductFromPortfolioInput{
		ProductId:   f.productID,
		PortfolioId: f.portfolioID,
	})

	return err
}

func (f *ServiceCatalogPortfolioProductAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PortfolioID", f.portfolioID)
	properties.Set("PortfolioName", f.portfolioName)
	properties.Set("ProductID", f.productID)
	properties.Set("ProductName", f.productName)
	return properties
}

func (f *ServiceCatalogPortfolioProductAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.productID, *f.portfolioID)
}
