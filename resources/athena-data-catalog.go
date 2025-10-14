package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/athena" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AthenaDataCatalogResource = "AthenaDataCatalog"

func init() {
	registry.Register(&registry.Registration{
		Name:     AthenaDataCatalogResource,
		Scope:    nuke.Account,
		Resource: &AthenaDataCatalog{},
		Lister:   &AthenaDataCatalogLister{},
	})
}

type AthenaDataCatalogLister struct{}

func (l *AthenaDataCatalogLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := athena.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &athena.ListDataCatalogsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListDataCatalogs(params)
		if err != nil {
			return nil, err
		}

		for _, catalog := range output.DataCatalogsSummary {
			resources = append(resources, &AthenaDataCatalog{
				svc:  svc,
				Name: catalog.CatalogName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type AthenaDataCatalog struct {
	svc  *athena.Athena
	Name *string
}

func (r *AthenaDataCatalog) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AthenaDataCatalog) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDataCatalog(&athena.DeleteDataCatalogInput{
		Name: r.Name,
	})

	return err
}

func (r *AthenaDataCatalog) Filter() error {
	if *r.Name == "AwsDataCatalog" {
		return fmt.Errorf("cannot delete default data source")
	}
	return nil
}

func (r *AthenaDataCatalog) String() string {
	return *r.Name
}
