package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/redshift" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RedshiftScheduledActionResource = "RedshiftScheduledAction"

func init() {
	registry.Register(&registry.Registration{
		Name:     RedshiftScheduledActionResource,
		Scope:    nuke.Account,
		Resource: &RedshiftScheduledAction{},
		Lister:   &RedshiftScheduledActionLister{},
	})
}

type RedshiftScheduledActionLister struct{}

func (l *RedshiftScheduledActionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshift.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshift.DescribeScheduledActionsInput{}

	for {
		resp, err := svc.DescribeScheduledActions(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.ScheduledActions {
			resources = append(resources, &RedshiftScheduledAction{
				svc:                 svc,
				scheduledActionName: item.ScheduledActionName,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type RedshiftScheduledAction struct {
	svc                 *redshift.Redshift
	scheduledActionName *string
}

func (f *RedshiftScheduledAction) Remove(_ context.Context) error {
	_, err := f.svc.DeleteScheduledAction(&redshift.DeleteScheduledActionInput{
		ScheduledActionName: f.scheduledActionName,
	})

	return err
}

func (f *RedshiftScheduledAction) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("scheduledActionName", f.scheduledActionName)
	return properties
}
