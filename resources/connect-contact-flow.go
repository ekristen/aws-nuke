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

const ConnectContactFlowResource = "ConnectContactFlow"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectContactFlowResource,
		Scope:    nuke.Account,
		Resource: &ConnectContactFlow{},
		Lister:   &ConnectContactFlowLister{},
	})
}

type ConnectContactFlowLister struct{}

func (l *ConnectContactFlowLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListContactFlowsPaginator(svc, &connect.ListContactFlowsInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, flow := range resp.ContactFlowSummaryList {
				var tags map[string]string
				if flow.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: flow.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect contact flow: %s", *flow.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectContactFlow{
					svc:              svc,
					InstanceID:       instance.Id,
					ContactFlowID:    flow.Id,
					Name:             flow.Name,
					ContactFlowType:  string(flow.ContactFlowType),
					ContactFlowState: string(flow.ContactFlowState),
					ARN:              flow.Arn,
					Tags:             tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectContactFlow struct {
	svc              *connect.Client
	InstanceID       *string
	ContactFlowID    *string
	Name             *string
	ContactFlowType  string
	ContactFlowState string
	ARN              *string
	Tags             map[string]string
}

func (r *ConnectContactFlow) Filter() error {
	// Only CONTACT_FLOW and CAMPAIGN types can be deleted; all other types are system-managed
	if r.ContactFlowType != string(connecttypes.ContactFlowTypeContactFlow) &&
		r.ContactFlowType != string(connecttypes.ContactFlowTypeCampaign) {
		return fmt.Errorf("cannot delete system-managed contact flow type: %s", r.ContactFlowType)
	}
	return nil
}

func (r *ConnectContactFlow) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteContactFlow(ctx, &connect.DeleteContactFlowInput{
		InstanceId:    r.InstanceID,
		ContactFlowId: r.ContactFlowID,
	})
	return err
}

func (r *ConnectContactFlow) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectContactFlow) String() string {
	return *r.Name
}
