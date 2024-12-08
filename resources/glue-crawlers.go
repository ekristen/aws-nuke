package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueCrawlerResource = "GlueCrawler"

func init() {
	registry.Register(&registry.Registration{
		Name:     GlueCrawlerResource,
		Scope:    nuke.Account,
		Resource: &GlueCrawler{},
		Lister:   &GlueCrawlerLister{},
	})
}

type GlueCrawlerLister struct{}

func (l *GlueCrawlerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetCrawlersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetCrawlers(params)
		if err != nil {
			return nil, err
		}

		for _, crawler := range output.Crawlers {
			resources = append(resources, &GlueCrawler{
				svc:  svc,
				name: crawler.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueCrawler struct {
	svc  *glue.Glue
	name *string
}

func (f *GlueCrawler) Remove(_ context.Context) error {
	_, err := f.svc.DeleteCrawler(&glue.DeleteCrawlerInput{
		Name: f.name,
	})

	return err
}

func (f *GlueCrawler) String() string {
	return *f.name
}
