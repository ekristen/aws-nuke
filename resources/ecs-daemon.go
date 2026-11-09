package resources

import (
	"context"
	"errors"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// ECSDaemonResource is the resource type identifier for ECS daemons. Daemons run
// agents on the instances of a capacity provider; while a daemon exists, its
// capacity provider (and therefore the cluster) cannot be deleted.
const ECSDaemonResource = "ECSDaemon"

func init() {
	registry.Register(&registry.Registration{
		Name:     ECSDaemonResource,
		Scope:    nuke.Account,
		Resource: &ECSDaemon{},
		Lister:   &ECSDaemonLister{},
	})
}

// ECSDaemonClient is the interface for the ECS operations used by ECSDaemon.
type ECSDaemonClient interface {
	ListDaemons(ctx context.Context, params *ecs.ListDaemonsInput,
		optFns ...func(*ecs.Options)) (*ecs.ListDaemonsOutput, error)
	ListTagsForResource(ctx context.Context, params *ecs.ListTagsForResourceInput,
		optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error)
	DeleteDaemon(ctx context.Context, params *ecs.DeleteDaemonInput,
		optFns ...func(*ecs.Options)) (*ecs.DeleteDaemonOutput, error)
}

type ECSDaemonLister struct {
	mockSvc ECSDaemonClient
}

func (l *ECSDaemonLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc ECSDaemonClient
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = ecs.NewFromConfig(*opts.Config)
	}

	resources := make([]resource.Resource, 0)

	// Without a cluster filter, ListDaemons returns the daemons of all clusters.
	params := &ecs.ListDaemonsInput{
		MaxResults: ptr.Int32(100),
	}

	for {
		output, err := svc.ListDaemons(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range output.DaemonSummariesList {
			daemon := &output.DaemonSummariesList[i]

			r := &ECSDaemon{
				svc:       svc,
				ARN:       daemon.DaemonArn,
				Status:    ptr.String(string(daemon.Status)),
				CreatedAt: daemon.CreatedAt,
				Tags:      make(map[string]string),
			}

			tags, err := svc.ListTagsForResource(ctx, &ecs.ListTagsForResourceInput{
				ResourceArn: daemon.DaemonArn,
			})
			if err != nil {
				logrus.WithError(err).Error("unable to get tags for ECS daemon")
			} else if tags != nil {
				for _, tag := range tags.Tags {
					if tag.Key != nil && tag.Value != nil {
						r.Tags[*tag.Key] = *tag.Value
					}
				}
			}

			resources = append(resources, r)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ECSDaemon struct {
	svc       ECSDaemonClient
	ARN       *string           `description:"The ARN of the ECS daemon"`
	Status    *string           `description:"The status of the ECS daemon"`
	CreatedAt *time.Time        `description:"The time the ECS daemon was created"`
	Tags      map[string]string `description:"The tags associated with the ECS daemon"`
}

func (r *ECSDaemon) Filter() error {
	if ptr.ToString(r.Status) == string(ecstypes.DaemonStatusDeleteInProgress) {
		return errors.New("already being deleted")
	}

	return nil
}

func (r *ECSDaemon) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDaemon(ctx, &ecs.DeleteDaemonInput{
		DaemonArn: r.ARN,
	})
	return err
}

func (r *ECSDaemon) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ECSDaemon) String() string {
	return ptr.ToString(r.ARN)
}
