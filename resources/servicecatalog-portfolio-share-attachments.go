package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ServiceCatalogPortfolioShareAttachmentResource = "ServiceCatalogPortfolioShareAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:   ServiceCatalogPortfolioShareAttachmentResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogPortfolioShareAttachmentLister{},
	})
}

type ServiceCatalogPortfolioShareAttachmentLister struct{}

func (l *ServiceCatalogPortfolioShareAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var portfolios []*servicecatalog.PortfolioDetail

	params := &servicecatalog.ListPortfoliosInput{
		PageSize: aws.Int64(20),
	}

	// list all portfolios
	for {
		resp, err := svc.ListPortfolios(params)
		if err != nil {
			return nil, err
		}

		portfolios = append(portfolios, resp.PortfolioDetails...)

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	accessParams := &servicecatalog.ListPortfolioAccessInput{}

	// Get all accounts which have shared access to the portfolio
	for _, portfolio := range portfolios {
		accessParams.PortfolioId = portfolio.Id

		resp, err := svc.ListPortfolioAccess(accessParams)
		if err != nil {
			return nil, err
		}

		for _, account := range resp.AccountIds {
			resources = append(resources, &ServiceCatalogPortfolioShareAttachment{
				svc:           svc,
				portfolioID:   portfolio.Id,
				accountID:     account,
				portfolioName: portfolio.DisplayName,
			})
		}
	}

	return resources, nil
}

type ServiceCatalogPortfolioShareAttachment struct {
	svc           *servicecatalog.ServiceCatalog
	portfolioID   *string
	accountID     *string
	portfolioName *string
}

func (f *ServiceCatalogPortfolioShareAttachment) Remove(_ context.Context) error {
	_, err := f.svc.DeletePortfolioShare(&servicecatalog.DeletePortfolioShareInput{
		AccountId:   f.accountID,
		PortfolioId: f.portfolioID,
	})

	return err
}

func (f *ServiceCatalogPortfolioShareAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PortfolioID", f.portfolioID)
	properties.Set("PortfolioName", f.portfolioName)
	properties.Set("AccountID", f.accountID)
	return properties
}

func (f *ServiceCatalogPortfolioShareAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.portfolioID, *f.accountID)
}
