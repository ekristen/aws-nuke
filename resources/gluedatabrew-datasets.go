package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gluedatabrew"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlueDataBrewDatasetsResource = "GlueDataBrewDatasets"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueDataBrewDatasetsResource,
		Scope:  nuke.Account,
		Lister: &GlueDataBrewDatasetsLister{},
	})
}

type GlueDataBrewDatasetsLister struct{}

func (l *GlueDataBrewDatasetsLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gluedatabrew.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &gluedatabrew.ListDatasetsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListDatasets(params)
		if err != nil {
			return nil, err
		}

		for _, dataset := range output.Datasets {
			resources = append(resources, &GlueDataBrewDatasets{
				svc:  svc,
				name: dataset.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDataBrewDatasets struct {
	svc  *gluedatabrew.GlueDataBrew
	name *string
}

func (f *GlueDataBrewDatasets) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDataset(&gluedatabrew.DeleteDatasetInput{
		Name: f.name,
	})

	return err
}

func (f *GlueDataBrewDatasets) String() string {
	return *f.name
}
