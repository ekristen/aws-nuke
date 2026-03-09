package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/connect"
	connecttypes "github.com/aws/aws-sdk-go-v2/service/connect/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectRoutingProfileResource = "ConnectRoutingProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectRoutingProfileResource,
		Scope:    nuke.Account,
		Resource: &ConnectRoutingProfile{},
		Lister:   &ConnectRoutingProfileLister{},
		DependsOn: []string{
			ConnectUserResource,
		},
	})
}

type ConnectRoutingProfileLister struct{}

func (l *ConnectRoutingProfileLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListRoutingProfilesPaginator(svc, &connect.ListRoutingProfilesInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, profile := range resp.RoutingProfileSummaryList {
				var tags map[string]string
				if profile.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: profile.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect routing profile: %s", *profile.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectRoutingProfile{
					svc:        svc,
					InstanceID: instance.Id,
					ProfileID:  profile.Id,
					Name:       profile.Name,
					ARN:        profile.Arn,
					Tags:       tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectRoutingProfile struct {
	svc        *connect.Client
	InstanceID *string
	ProfileID  *string
	Name       *string
	ARN        *string
	Tags       map[string]string
}

func (r *ConnectRoutingProfile) Filter() error {
	if r.Name != nil && *r.Name == "Basic Routing Profile" {
		return fmt.Errorf("cannot delete default routing profile")
	}
	return nil
}

func (r *ConnectRoutingProfile) Remove(ctx context.Context) error {
	// First, disassociate all queues from the routing profile
	queuePaginator := connect.NewListRoutingProfileQueuesPaginator(r.svc, &connect.ListRoutingProfileQueuesInput{
		InstanceId:       r.InstanceID,
		RoutingProfileId: r.ProfileID,
	})

	for queuePaginator.HasMorePages() {
		resp, err := queuePaginator.NextPage(ctx)
		if err != nil {
			return err
		}

		// Batch disassociate in groups of 10
		var refs []connecttypes.RoutingProfileQueueReference
		for _, q := range resp.RoutingProfileQueueConfigSummaryList {
			refs = append(refs, connecttypes.RoutingProfileQueueReference{
				QueueId: q.QueueId,
				Channel: q.Channel,
			})

			if len(refs) == 10 {
				_, err := r.svc.DisassociateRoutingProfileQueues(ctx, &connect.DisassociateRoutingProfileQueuesInput{
					InstanceId:       r.InstanceID,
					RoutingProfileId: r.ProfileID,
					QueueReferences:  refs,
				})
				if err != nil {
					return err
				}
				refs = nil
			}
		}

		if len(refs) > 0 {
			_, err := r.svc.DisassociateRoutingProfileQueues(ctx, &connect.DisassociateRoutingProfileQueuesInput{
				InstanceId:       r.InstanceID,
				RoutingProfileId: r.ProfileID,
				QueueReferences:  refs,
			})
			if err != nil {
				return err
			}
		}
	}

	_, err := r.svc.DeleteRoutingProfile(ctx, &connect.DeleteRoutingProfileInput{
		InstanceId:       r.InstanceID,
		RoutingProfileId: r.ProfileID,
	})
	return err
}

func (r *ConnectRoutingProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectRoutingProfile) String() string {
	return *r.Name
}
