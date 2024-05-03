package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockModelInvocationLoggingConfigurationResource = "BedrockModelInvocationLoggingConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockModelInvocationLoggingConfigurationResource,
		Scope:  nuke.Account,
		Lister: &BedrockModelInvocationLoggingConfigurationLister{},
	})
}

type BedrockModelInvocationLoggingConfigurationLister struct{}

func (l *BedrockModelInvocationLoggingConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// There is nothing to "list" because there is only one logging configuration per account and deleting it require no params
	resp, err := svc.GetModelInvocationLoggingConfiguration(&bedrock.GetModelInvocationLoggingConfigurationInput{})
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.LoggingConfig != nil {
		resources = append(resources, &BedrockModelInvocationLoggingConfiguration{
			svc: svc,
		})
	}

	return resources, nil
}

type BedrockModelInvocationLoggingConfiguration struct {
	svc *bedrock.Bedrock
}

func (r *BedrockModelInvocationLoggingConfiguration) Remove(_ context.Context) error {
	_, err := r.svc.DeleteModelInvocationLoggingConfiguration(&bedrock.DeleteModelInvocationLoggingConfigurationInput{})

	return err
}

func (r *BedrockModelInvocationLoggingConfiguration) String() string {
	return awsutil.Default
}

func (r *BedrockModelInvocationLoggingConfiguration) Properties() types.Properties {
	return types.NewProperties()
}
