package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"               //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/qbusiness" //nolint:staticcheck

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

func (l *QBusinessPluginLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := qbusiness.New(opts.Session)
	resources := make([]resource.Resource, 0)

	apps, err := listQBusinessApplications(svc)
	if err != nil {
		return nil, err
	}

	for _, appID := range apps {
		params := &qbusiness.ListPluginsInput{
			ApplicationId: appID,
			MaxResults:    aws.Int64(50),
		}
		for {
			resp, err := svc.ListPlugins(params)
			if err != nil {
				return nil, err
			}
			for _, p := range resp.Plugins {
				resources = append(resources, &QBusinessPlugin{
					svc:           svc,
					ApplicationID: appID,
					ID:            p.PluginId,
					Name:          p.DisplayName,
					State:         p.State,
				})
			}
			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type QBusinessPlugin struct {
	svc           *qbusiness.QBusiness
	ApplicationID *string
	ID            *string
	Name          *string
	State         *string
}

func (r *QBusinessPlugin) Remove(_ context.Context) error {
	_, err := r.svc.DeletePlugin(&qbusiness.DeletePluginInput{
		ApplicationId: r.ApplicationID,
		PluginId:      r.ID,
	})
	return err
}

func (r *QBusinessPlugin) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QBusinessPlugin) String() string {
	return aws.StringValue(r.ID)
}
