package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ModelInvocationLoggingConfigurationResource = "ModelInvocationLoggingConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   ModelInvocationLoggingConfigurationResource,
		Scope:  nuke.Account,
		Lister: &ModelInvocationLoggingConfigurationLister{},
	})
}

type ModelInvocationLoggingConfigurationLister struct{}

func (l *ModelInvocationLoggingConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// There is nothing to "list" because there is only one logging configuration per account and deleting it require no params
	resources = append(resources, &ModelInvocationLoggingConfiguration{
		svc: svc,
	})

	return resources, nil
}

type ModelInvocationLoggingConfiguration struct {
	svc *bedrock.Bedrock
}

func (f *ModelInvocationLoggingConfiguration) Remove(_ context.Context) error {
	_, err := f.svc.DeleteModelInvocationLoggingConfiguration(&bedrock.DeleteModelInvocationLoggingConfigurationInput{})

	return err
}

func (f *ModelInvocationLoggingConfiguration) String() string {
	return "default"
}

func (f *ModelInvocationLoggingConfiguration) Properties() types.Properties {
	properties := types.NewProperties()
	return properties
}
