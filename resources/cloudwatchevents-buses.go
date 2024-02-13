package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudwatchevents"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudWatchEventsBusesResource = "CloudWatchEventsBuses"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudWatchEventsBusesResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchEventsBusesLister{},
	})
}

type CloudWatchEventsBusesLister struct{}

func (l *CloudWatchEventsBusesLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchevents.New(opts.Session)

	resp, err := svc.ListEventBuses(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, bus := range resp.EventBuses {
		if *bus.Name == "default" {
			continue
		}

		resources = append(resources, &CloudWatchEventsBus{
			svc:  svc,
			name: bus.Name,
		})
	}
	return resources, nil
}

type CloudWatchEventsBus struct {
	svc  *cloudwatchevents.CloudWatchEvents
	name *string
}

func (bus *CloudWatchEventsBus) Remove(_ context.Context) error {
	_, err := bus.svc.DeleteEventBus(&cloudwatchevents.DeleteEventBusInput{
		Name: bus.name,
	})
	return err
}

func (bus *CloudWatchEventsBus) String() string {
	return *bus.name
}
