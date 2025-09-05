package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mgn"
	"github.com/aws/aws-sdk-go-v2/service/mgn/types"
	"github.com/aws/smithy-go"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	MGNSourceServerResource                      = "MGNSourceServer"
	mgnSourceServerUninitializedAccountException = "UninitializedAccountException"
)

func init() {
	registry.Register(&registry.Registration{
		Name:     MGNSourceServerResource,
		Scope:    nuke.Account,
		Resource: &MGNSourceServer{},
		Lister:   &MGNSourceServerLister{},
	})
}

type MGNSourceServerLister struct{}

func (l *MGNSourceServerLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &mgn.DescribeSourceServersInput{
		MaxResults: aws.Int32(50),
	}

	for {
		output, err := svc.DescribeSourceServers(ctx, params)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == mgnSourceServerUninitializedAccountException {
				return nil, nil
			}
			return nil, err
		}

		for i := range output.Items {
			sourceServer := &output.Items[i]
			mgnServer := &MGNSourceServer{
				svc:             svc,
				sourceServer:    sourceServer,
				SourceServerID:  sourceServer.SourceServerID,
				Arn:             sourceServer.Arn,
				ReplicationType: string(sourceServer.ReplicationType),
				IsArchived:      sourceServer.IsArchived,
				Tags:            sourceServer.Tags,
			}

			if sourceServer.LifeCycle != nil {
				mgnServer.LifeCycleState = string(sourceServer.LifeCycle.State)
			}

			if sourceServer.SourceProperties != nil && sourceServer.SourceProperties.IdentificationHints != nil {
				mgnServer.Hostname = sourceServer.SourceProperties.IdentificationHints.Hostname
				mgnServer.FQDN = sourceServer.SourceProperties.IdentificationHints.Fqdn
			}

			resources = append(resources, mgnServer)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNSourceServer struct {
	svc          *mgn.Client         `description:"-"`
	sourceServer *types.SourceServer `description:"-"`

	// Exposed properties
	SourceServerID  *string           `description:"The unique identifier of the source server"`
	Arn             *string           `description:"The ARN of the source server"`
	ReplicationType string            `description:"The type of replication (AGENT_BASED, etc.)"`
	IsArchived      *bool             `description:"Whether the source server is archived"`
	LifeCycleState  string            `description:"The lifecycle state of the source server"`
	Hostname        *string           `description:"The hostname of the source server"`
	FQDN            *string           `description:"The fully qualified domain name of the source server"`
	Tags            map[string]string `description:"The tags associated with the source server"`
}

func (f *MGNSourceServer) Remove(ctx context.Context) error {
	// Disconnect source server from service first before delete
	if _, err := f.svc.DisconnectFromService(ctx, &mgn.DisconnectFromServiceInput{
		SourceServerID: f.sourceServer.SourceServerID,
	}); err != nil {
		return err
	}

	_, err := f.svc.DeleteSourceServer(ctx, &mgn.DeleteSourceServerInput{
		SourceServerID: f.sourceServer.SourceServerID,
	})

	return err
}

func (f *MGNSourceServer) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(f)
}

func (f *MGNSourceServer) String() string {
	return *f.SourceServerID
}
