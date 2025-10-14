package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"                    //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/servicecatalog" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ServiceCatalogPrincipalPortfolioAttachmentResource = "ServiceCatalogPrincipalPortfolioAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     ServiceCatalogPrincipalPortfolioAttachmentResource,
		Scope:    nuke.Account,
		Resource: &ServiceCatalogPrincipalPortfolioAttachment{},
		Lister:   &ServiceCatalogPrincipalPortfolioAttachmentLister{},
	})
}

type ServiceCatalogPrincipalPortfolioAttachmentLister struct{}

func (l *ServiceCatalogPrincipalPortfolioAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var portfolios []*servicecatalog.PortfolioDetail

	params := &servicecatalog.ListPortfoliosInput{
		PageSize: aws.Int64(20),
	}

	// List all Portfolios
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

	principalParams := &servicecatalog.ListPrincipalsForPortfolioInput{
		PageSize: aws.Int64(20),
	}

	for _, portfolio := range portfolios {
		principalParams.PortfolioId = portfolio.Id

		resp, err := svc.ListPrincipalsForPortfolio(principalParams)
		if err != nil {
			return nil, err
		}

		for _, principal := range resp.Principals {
			resources = append(resources, &ServiceCatalogPrincipalPortfolioAttachment{
				svc:           svc,
				principalARN:  principal.PrincipalARN,
				portfolioID:   portfolio.Id,
				portfolioName: portfolio.DisplayName,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		principalParams.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ServiceCatalogPrincipalPortfolioAttachment struct {
	svc           *servicecatalog.ServiceCatalog
	portfolioID   *string
	principalARN  *string
	portfolioName *string
}

func (f *ServiceCatalogPrincipalPortfolioAttachment) Remove(_ context.Context) error {
	_, err := f.svc.DisassociatePrincipalFromPortfolio(&servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PrincipalARN: f.principalARN,
		PortfolioId:  f.portfolioID,
	})

	return err
}

func (f *ServiceCatalogPrincipalPortfolioAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PortfolioID", f.portfolioID)
	properties.Set("PrincipalARN", f.principalARN)
	properties.Set("PortfolioName", f.portfolioName)
	return properties
}

func (f *ServiceCatalogPrincipalPortfolioAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.principalARN, *f.portfolioID)
}
