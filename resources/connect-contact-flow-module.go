package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectContactFlowModuleResource = "ConnectContactFlowModule"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectContactFlowModuleResource,
		Scope:    nuke.Account,
		Resource: &ConnectContactFlowModule{},
		Lister:   &ConnectContactFlowModuleLister{},
	})
}

type ConnectContactFlowModuleLister struct{}

func (l *ConnectContactFlowModuleLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListContactFlowModulesPaginator(svc, &connect.ListContactFlowModulesInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, module := range resp.ContactFlowModulesSummaryList {
				var tags map[string]string
				if module.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: module.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect contact flow module: %s", *module.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectContactFlowModule{
					svc:        svc,
					InstanceID: instance.Id,
					ModuleID:   module.Id,
					Name:       module.Name,
					State:      string(module.State),
					ARN:        module.Arn,
					Tags:       tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectContactFlowModule struct {
	svc        *connect.Client
	InstanceID *string
	ModuleID   *string
	Name       *string
	State      string
	ARN        *string
	Tags       map[string]string
}

func (r *ConnectContactFlowModule) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteContactFlowModule(ctx, &connect.DeleteContactFlowModuleInput{
		InstanceId:          r.InstanceID,
		ContactFlowModuleId: r.ModuleID,
	})
	return err
}

func (r *ConnectContactFlowModule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectContactFlowModule) String() string {
	return *r.Name
}
