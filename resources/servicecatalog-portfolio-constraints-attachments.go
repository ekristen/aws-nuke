package resources

import (
	"context"

	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ServiceCatalogConstraintPortfolioAttachmentResource = "ServiceCatalogConstraintPortfolioAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:   ServiceCatalogConstraintPortfolioAttachmentResource,
		Scope:  nuke.Account,
		Lister: &ServiceCatalogConstraintPortfolioAttachmentLister{},
	})
}

type ServiceCatalogConstraintPortfolioAttachmentLister struct{}

func (l *ServiceCatalogConstraintPortfolioAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := servicecatalog.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var portfolios []*servicecatalog.PortfolioDetail

	params := &servicecatalog.ListPortfoliosInput{
		PageSize: aws.Int64(20),
	}

	//List all Portfolios
	for {
		resp, err := svc.ListPortfolios(params)
		if err != nil {
			if awsutil.IsAWSError(err, servicecatalog.ErrCodeTagOptionNotMigratedException) {
				logrus.Info(err)
				break
			}
			return nil, err
		}

		for _, portfolioDetail := range resp.PortfolioDetails {
			portfolios = append(portfolios, portfolioDetail)
		}

		if resp.NextPageToken == nil {
			break
		}

		params.PageToken = resp.NextPageToken
	}

	constraintParams := &servicecatalog.ListConstraintsForPortfolioInput{
		PageSize: aws.Int64(20),
	}

	for _, portfolio := range portfolios {

		constraintParams.PortfolioId = portfolio.Id
		resp, err := svc.ListConstraintsForPortfolio(constraintParams)
		if err != nil {
			return nil, err
		}

		for _, constraintDetail := range resp.ConstraintDetails {
			resources = append(resources, &ServiceCatalogConstraintPortfolioAttachment{
				svc:           svc,
				portfolioID:   portfolio.Id,
				constraintID:  constraintDetail.ConstraintId,
				portfolioName: portfolio.DisplayName,
			})
		}

		if resp.NextPageToken == nil {
			break
		}

		constraintParams.PageToken = resp.NextPageToken
	}

	return resources, nil
}

type ServiceCatalogConstraintPortfolioAttachment struct {
	svc           *servicecatalog.ServiceCatalog
	constraintID  *string
	portfolioID   *string
	portfolioName *string
}

func (f *ServiceCatalogConstraintPortfolioAttachment) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConstraint(&servicecatalog.DeleteConstraintInput{
		Id: f.constraintID,
	})

	return err
}

func (f *ServiceCatalogConstraintPortfolioAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PortfolioID", f.portfolioID)
	properties.Set("ConstraintID", f.constraintID)
	properties.Set("PortfolioName", f.portfolioName)
	return properties
}

func (f *ServiceCatalogConstraintPortfolioAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.constraintID, *f.portfolioID)
}
