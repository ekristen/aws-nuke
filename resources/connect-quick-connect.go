package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectQuickConnectResource = "ConnectQuickConnect"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectQuickConnectResource,
		Scope:    nuke.Account,
		Resource: &ConnectQuickConnect{},
		Lister:   &ConnectQuickConnectLister{},
	})
}

type ConnectQuickConnectLister struct{}

func (l *ConnectQuickConnectLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListQuickConnectsPaginator(svc, &connect.ListQuickConnectsInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, qc := range resp.QuickConnectSummaryList {
				var tags map[string]string
				if qc.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: qc.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect quick connect: %s", *qc.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectQuickConnect{
					svc:              svc,
					InstanceID:       instance.Id,
					QuickConnectID:   qc.Id,
					Name:             qc.Name,
					QuickConnectType: string(qc.QuickConnectType),
					ARN:              qc.Arn,
					Tags:             tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectQuickConnect struct {
	svc              *connect.Client
	InstanceID       *string
	QuickConnectID   *string
	Name             *string
	QuickConnectType string
	ARN              *string
	Tags             map[string]string
}

func (r *ConnectQuickConnect) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteQuickConnect(ctx, &connect.DeleteQuickConnectInput{
		InstanceId:     r.InstanceID,
		QuickConnectId: r.QuickConnectID,
	})
	return err
}

func (r *ConnectQuickConnect) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectQuickConnect) String() string {
	return *r.Name
}
