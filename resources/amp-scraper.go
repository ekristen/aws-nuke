package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/amp"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AMPScraperResource = "AMPScraper"

func init() {
	registry.Register(&registry.Registration{
		Name:     AMPScraperResource,
		Scope:    nuke.Account,
		Resource: &AMPScraper{},
		Lister:   &AMPScraperLister{},
	})
}

type AMPScraperLister struct{}

func (l *AMPScraperLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := amp.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := amp.NewListScrapersPaginator(svc, &amp.ListScrapersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, ws := range page.Scrapers {
			resources = append(resources, &AMPScraper{
				svc:       svc,
				ScraperID: ws.ScraperId,
				Alias:     ws.Alias,
				Tags:      ws.Tags,
			})
		}
	}

	return resources, nil
}

type AMPScraper struct {
	svc       *amp.Client
	ScraperID *string           `description:"The ID of the AMP Scraper"`
	Alias     *string           `description:"The alias of the AMP Scraper"`
	Tags      map[string]string `description:"The tags of the AMP Scraper"`
}

func (f *AMPScraper) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteScraper(ctx, &amp.DeleteScraperInput{
		ScraperId: f.ScraperID,
	})

	return err
}

func (f *AMPScraper) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}
