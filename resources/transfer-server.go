package resources

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"              //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/transfer" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const TransferServerResource = "TransferServer"

func init() {
	registry.Register(&registry.Registration{
		Name:     TransferServerResource,
		Scope:    nuke.Account,
		Resource: &TransferServer{},
		Lister:   &TransferServerLister{},
	})
}

type TransferServerLister struct{}

func (l *TransferServerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transfer.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &transfer.ListServersInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListServers(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Servers {
			descOutput, err := svc.DescribeServer(&transfer.DescribeServerInput{
				ServerId: item.ServerId,
			})
			if err != nil {
				return nil, err
			}

			var protocols []string
			for _, protocol := range descOutput.Server.Protocols {
				protocols = append(protocols, *protocol)
			}

			resources = append(resources, &TransferServer{
				svc:          svc,
				serverID:     item.ServerId,
				endpointType: item.EndpointType,
				protocols:    protocols,
				tags:         descOutput.Server.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type TransferServer struct {
	svc          *transfer.Transfer
	serverID     *string
	endpointType *string
	protocols    []string
	tags         []*transfer.Tag
}

func (ts *TransferServer) Remove(_ context.Context) error {
	_, err := ts.svc.DeleteServer(&transfer.DeleteServerInput{
		ServerId: ts.serverID,
	})

	return err
}

func (ts *TransferServer) String() string {
	return *ts.serverID
}

func (ts *TransferServer) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range ts.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.
		Set("ServerID", ts.serverID).
		Set("EndpointType", ts.endpointType).
		Set("Protocols", strings.Join(ts.protocols, ", "))
	return properties
}
