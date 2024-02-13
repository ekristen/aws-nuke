package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/kafka"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MSKConfigurationResource = "MSKConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   MSKConfigurationResource,
		Scope:  nuke.Account,
		Lister: &MSKConfigurationLister{},
	})
}

type MSKConfigurationLister struct{}

func (l *MSKConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kafka.New(opts.Session)
	params := &kafka.ListConfigurationsInput{}
	resp, err := svc.ListConfigurations(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, configuration := range resp.Configurations {
		resources = append(resources, &MSKConfiguration{
			svc:  svc,
			arn:  *configuration.Arn,
			name: *configuration.Name,
		})
	}

	return resources, nil
}

type MSKConfiguration struct {
	svc  *kafka.Kafka
	arn  string
	name string
}

func (m *MSKConfiguration) Remove(_ context.Context) error {
	params := &kafka.DeleteConfigurationInput{
		Arn: &m.arn,
	}

	_, err := m.svc.DeleteConfiguration(params)
	if err != nil {
		return err
	}

	return nil
}

func (m *MSKConfiguration) String() string {
	return m.arn
}

func (m *MSKConfiguration) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", m.arn)
	properties.Set("Name", m.name)

	return properties
}
