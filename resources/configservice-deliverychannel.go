package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/configservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConfigServiceDeliveryChannelResource = "ConfigServiceDeliveryChannel"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConfigServiceDeliveryChannelResource,
		Scope:    nuke.Account,
		Resource: &ConfigServiceDeliveryChannel{},
		Lister:   &ConfigServiceDeliveryChannelLister{},
	})
}

type ConfigServiceDeliveryChannelLister struct{}

func (l *ConfigServiceDeliveryChannelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)
	svc := configservice.New(opts.Session)

	params := &configservice.DescribeDeliveryChannelsInput{}
	resp, err := svc.DescribeDeliveryChannels(params)
	if err != nil {
		return nil, err
	}

	for _, deliveryChannel := range resp.DeliveryChannels {
		resources = append(resources, &ConfigServiceDeliveryChannel{
			svc:  svc,
			Name: deliveryChannel.Name,
		})
	}

	return resources, nil
}

type ConfigServiceDeliveryChannel struct {
	svc  *configservice.ConfigService
	Name *string
}

func (r *ConfigServiceDeliveryChannel) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDeliveryChannel(&configservice.DeleteDeliveryChannelInput{
		DeliveryChannelName: r.Name,
	})

	return err
}

func (r *ConfigServiceDeliveryChannel) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConfigServiceDeliveryChannel) String() string {
	return *r.Name
}
