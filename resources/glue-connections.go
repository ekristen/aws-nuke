package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlueConnectionResource = "GlueConnection"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueConnectionResource,
		Scope:  nuke.Account,
		Lister: &GlueConnectionLister{},
	})
}

type GlueConnectionLister struct{}

func (l *GlueConnectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetConnectionsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetConnections(params)
		if err != nil {
			return nil, err
		}

		for _, connection := range output.ConnectionList {
			resources = append(resources, &GlueConnection{
				svc:            svc,
				connectionName: connection.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueConnection struct {
	svc            *glue.Glue
	connectionName *string
}

func (f *GlueConnection) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConnection(&glue.DeleteConnectionInput{
		ConnectionName: f.connectionName,
	})

	return err
}

func (f *GlueConnection) String() string {
	return *f.connectionName
}
