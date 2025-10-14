package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/configservice" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConfigServiceConfigurationRecorderResource = "ConfigServiceConfigurationRecorder"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConfigServiceConfigurationRecorderResource,
		Scope:    nuke.Account,
		Resource: &ConfigServiceConfigurationRecorder{},
		Lister:   &ConfigServiceConfigurationRecorderLister{},
	})
}

type ConfigServiceConfigurationRecorderLister struct{}

func (l *ConfigServiceConfigurationRecorderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := configservice.New(opts.Session)

	params := &configservice.DescribeConfigurationRecordersInput{}
	resp, err := svc.DescribeConfigurationRecorders(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, configurationRecorder := range resp.ConfigurationRecorders {
		resources = append(resources, &ConfigServiceConfigurationRecorder{
			svc:  svc,
			Name: configurationRecorder.Name,
		})
	}

	return resources, nil
}

type ConfigServiceConfigurationRecorder struct {
	svc  *configservice.ConfigService
	Name *string
}

func (r *ConfigServiceConfigurationRecorder) Remove(_ context.Context) error {
	_, err := r.svc.DeleteConfigurationRecorder(&configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: r.Name,
	})

	return err
}

func (r *ConfigServiceConfigurationRecorder) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConfigServiceConfigurationRecorder) String() string {
	return *r.Name
}
