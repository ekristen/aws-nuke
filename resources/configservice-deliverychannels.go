package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/configservice"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ConfigServiceDeliveryChannelResource = "ConfigServiceDeliveryChannel"

func init() {
	registry.Register(&registry.Registration{
		Name:   ConfigServiceDeliveryChannelResource,
		Scope:  nuke.Account,
		Lister: &ConfigServiceDeliveryChannelLister{},
	})
}

type ConfigServiceDeliveryChannelLister struct{}

func (l *ConfigServiceDeliveryChannelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := configservice.New(opts.Session)

	params := &configservice.DescribeDeliveryChannelsInput{}
	resp, err := svc.DescribeDeliveryChannels(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, deliveryChannel := range resp.DeliveryChannels {
		resources = append(resources, &ConfigServiceDeliveryChannel{
			svc:                 svc,
			deliveryChannelName: deliveryChannel.Name,
		})
	}

	return resources, nil
}

type ConfigServiceDeliveryChannel struct {
	svc                 *configservice.ConfigService
	deliveryChannelName *string
}

func (f *ConfigServiceDeliveryChannel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDeliveryChannel(&configservice.DeleteDeliveryChannelInput{
		DeliveryChannelName: f.deliveryChannelName,
	})

	return err
}

func (f *ConfigServiceDeliveryChannel) String() string {
	return *f.deliveryChannelName
}
