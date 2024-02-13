package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/clouddirectory"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudDirectorySchemaResource = "CloudDirectorySchema"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudDirectorySchemaResource,
		Scope:  nuke.Account,
		Lister: &CloudDirectorySchemaLister{},
	})
}

type CloudDirectorySchemaLister struct{}

func (l *CloudDirectorySchemaLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := clouddirectory.New(opts.Session)
	resources := make([]resource.Resource, 0)

	developmentParams := &clouddirectory.ListDevelopmentSchemaArnsInput{
		MaxResults: aws.Int64(30),
	}

	// Get all development schemas
	for {
		resp, err := svc.ListDevelopmentSchemaArns(developmentParams)
		if err != nil {
			return nil, err
		}

		for _, arn := range resp.SchemaArns {
			resources = append(resources, &CloudDirectorySchema{
				svc:       svc,
				schemaARN: arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		developmentParams.NextToken = resp.NextToken
	}

	// Get all published schemas
	publishedParams := &clouddirectory.ListPublishedSchemaArnsInput{
		MaxResults: aws.Int64(30),
	}
	for {
		resp, err := svc.ListPublishedSchemaArns(publishedParams)
		if err != nil {
			return nil, err
		}

		for _, arn := range resp.SchemaArns {
			resources = append(resources, &CloudDirectorySchema{
				svc:       svc,
				schemaARN: arn,
			})
		}

		if resp.NextToken == nil {
			break
		}

		publishedParams.NextToken = resp.NextToken
	}

	// Return combined development and production schemas to DeleteSchema
	return resources, nil
}

type CloudDirectorySchema struct {
	svc       *clouddirectory.CloudDirectory
	schemaARN *string
}

func (f *CloudDirectorySchema) Remove(_ context.Context) error {

	_, err := f.svc.DeleteSchema(&clouddirectory.DeleteSchemaInput{
		SchemaArn: f.schemaARN,
	})

	return err
}

func (f *CloudDirectorySchema) String() string {
	return *f.schemaARN
}
