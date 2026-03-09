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

const ConnectQueueResource = "ConnectQueue"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectQueueResource,
		Scope:    nuke.Account,
		Resource: &ConnectQueue{},
		Lister:   &ConnectQueueLister{},
	})
}

type ConnectQueueLister struct{}

func (l *ConnectQueueLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListQueuesPaginator(svc, &connect.ListQueuesInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, queue := range resp.QueueSummaryList {
				var tags map[string]string
				if queue.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: queue.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect queue: %s", *queue.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectQueue{
					svc:        svc,
					InstanceID: instance.Id,
					QueueID:    queue.Id,
					Name:       queue.Name,
					QueueType:  string(queue.QueueType),
					ARN:        queue.Arn,
					Tags:       tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectQueue struct {
	svc        *connect.Client
	InstanceID *string
	QueueID    *string
	Name       *string
	QueueType  string
	ARN        *string
	Tags       map[string]string
}

func (r *ConnectQueue) Filter() error {
	if r.QueueType == string(connecttypes.QueueTypeAgent) {
		return fmt.Errorf("cannot delete agent queue (auto-created)")
	}
	return nil
}

func (r *ConnectQueue) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteQueue(ctx, &connect.DeleteQueueInput{
		InstanceId: r.InstanceID,
		QueueId:    r.QueueID,
	})
	return err
}

func (r *ConnectQueue) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectQueue) String() string {
	return *r.Name
}
