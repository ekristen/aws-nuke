package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeStarConnectionResource = "CodeStarConnection"

func init() {
	registry.Register(&registry.Registration{
		Name:   CodeStarConnectionResource,
		Scope:  nuke.Account,
		Lister: &CodeStarConnectionLister{},
	})
}

type CodeStarConnectionLister struct{}

func (l *CodeStarConnectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codestarconnections.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codestarconnections.ListConnectionsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListConnections(params)
		if err != nil {
			return nil, err
		}

		for _, connection := range output.Connections {
			resources = append(resources, &CodeStarConnection{
				svc:            svc,
				connectionARN:  connection.ConnectionArn,
				connectionName: connection.ConnectionName,
				providerType:   connection.ProviderType,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CodeStarConnection struct {
	svc            *codestarconnections.CodeStarConnections
	connectionARN  *string
	connectionName *string
	providerType   *string
}

func (f *CodeStarConnection) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConnection(&codestarconnections.DeleteConnectionInput{
		ConnectionArn: f.connectionARN,
	})

	return err
}

func (f *CodeStarConnection) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("Name", f.connectionName).
		Set("ProviderType", f.providerType)
	return properties
}

func (f *CodeStarConnection) String() string {
	return *f.connectionName
}
