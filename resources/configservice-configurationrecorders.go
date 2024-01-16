package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/configservice"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ConfigServiceConfigurationRecorderResource = "ConfigServiceConfigurationRecorder"

func init() {
	resource.Register(resource.Registration{
		Name:   ConfigServiceConfigurationRecorderResource,
		Scope:  nuke.Account,
		Lister: &ConfigServiceConfigurationRecorderLister{},
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
			svc:                       svc,
			configurationRecorderName: configurationRecorder.Name,
		})
	}

	return resources, nil
}

type ConfigServiceConfigurationRecorder struct {
	svc                       *configservice.ConfigService
	configurationRecorderName *string
}

func (f *ConfigServiceConfigurationRecorder) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConfigurationRecorder(&configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: f.configurationRecorderName,
	})

	return err
}

func (f *ConfigServiceConfigurationRecorder) String() string {
	return *f.configurationRecorderName
}
