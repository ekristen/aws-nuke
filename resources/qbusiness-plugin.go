package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const QBusinessPluginResource = "QBusinessPlugin"

func init() {
	registry.Register(&registry.Registration{
		Name:     QBusinessPluginResource,
		Scope:    nuke.Account,
		Resource: &QBusinessPlugin{},
		Lister:   &QBusinessPluginLister{},
	})
}

type QBusinessPluginLister struct{}

func (l *QBusinessPluginLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	appIDs, err := listQBusinessApplicationIDs(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range appIDs {
		paginator := qbusiness.NewListPluginsPaginator(svc, &qbusiness.ListPluginsInput{
			ApplicationId: appID,
			MaxResults:    aws.Int32(50),
		})
		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			for _, p := range resp.Plugins {
				resources = append(resources, &QBusinessPlugin{
					svc:           svc,
					ApplicationID: appID,
					ID:            p.PluginId,
					Name:          p.DisplayName,
					State:         aws.String(string(p.State)),
				})
			}
		}
	}

	return resources, nil
}

type QBusinessPlugin struct {
	svc           *qbusiness.Client
	ApplicationID *string
	ID            *string
	Name          *string
	State         *string
}

func (r *QBusinessPlugin) Remove(ctx context.Context) error {
	_, err := r.svc.DeletePlugin(ctx, &qbusiness.DeletePluginInput{
		ApplicationId: r.ApplicationID,
		PluginId:      r.ID,
	})
	return err
}

func (r *QBusinessPlugin) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessPlugin) String() string {
	return aws.ToString(r.ID)
}
