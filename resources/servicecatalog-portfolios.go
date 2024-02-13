package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogPortfolioResource = "ServiceCatalogPortfolio"

func init() {
	registry.Register(&registry.Registration{
		Name:   ServiceCatalogPortfolioResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogPortfolioLister{},
	})
}

type ServiceCatalogPortfolioLister struct{}

func (l *ServiceCatalogPortfolioLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &servicecatalog.ListPortfoliosInput{
		PageSize: aws.Int64(20),
	}

	for {
		resp, err := svc.ListPortfolios(params)
		if err != nil {
			return nil, err
		}

		for _, portfolioDetail := range resp.PortfolioDetails {
			resources = append(resources, &ServiceCatalogPortfolio{
				svc:          svc,
				ID:           portfolioDetail.Id,
				displayName:  portfolioDetail.DisplayName,
				providerName: portfolioDetail.ProviderName,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ServiceCatalogPortfolio struct {
	svc          *servicecatalog.ServiceCatalog
	ID           *string
	displayName  *string
	providerName *string
}

func (f *ServiceCatalogPortfolio) Remove(_ context.Context) error {
	_, err := f.svc.DeletePortfolio(&servicecatalog.DeletePortfolioInput{
		Id: f.ID,
	})

	return err
}

func (f *ServiceCatalogPortfolio) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("DisplayName", f.displayName)
	properties.Set("ProviderName", f.providerName)
	return properties
}

func (f *ServiceCatalogPortfolio) String() string {
	return *f.ID
}
