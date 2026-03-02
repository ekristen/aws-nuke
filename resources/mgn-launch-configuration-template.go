package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mgn"
	"github.com/aws/aws-sdk-go-v2/service/mgn/types"
	"github.com/aws/smithy-go"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	MGNLaunchConfigurationTemplateResource                      = "MGNLaunchConfigurationTemplate"
	mgnLaunchConfigurationTemplateUninitializedAccountException = "UninitializedAccountException"
)

func init() {
	registry.Register(&registry.Registration{
		Name:     MGNLaunchConfigurationTemplateResource,
		Scope:    nuke.Account,
		Resource: &MGNLaunchConfigurationTemplate{},
		Lister:   &MGNLaunchConfigurationTemplateLister{},
	})
}

type MGNLaunchConfigurationTemplateLister struct{}

func (l *MGNLaunchConfigurationTemplateLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &mgn.DescribeLaunchConfigurationTemplatesInput{
		MaxResults: aws.Int32(50),
	}

	for {
		output, err := svc.DescribeLaunchConfigurationTemplates(ctx, params)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == mgnLaunchConfigurationTemplateUninitializedAccountException {
				return nil, nil
			}
			return nil, err
		}

		for i := range output.Items {
			template := &output.Items[i]
			mgnTemplate := &MGNLaunchConfigurationTemplate{
				svc:                                 svc,
				template:                            template,
				LaunchConfigurationTemplateID:       template.LaunchConfigurationTemplateID,
				Arn:                                 template.Arn,
				Ec2LaunchTemplateID:                 template.Ec2LaunchTemplateID,
				LaunchDisposition:                   string(template.LaunchDisposition),
				TargetInstanceTypeRightSizingMethod: string(template.TargetInstanceTypeRightSizingMethod),
				CopyPrivateIP:                       template.CopyPrivateIp,
				CopyTags:                            template.CopyTags,
				EnableMapAutoTagging:                template.EnableMapAutoTagging,
				Tags:                                template.Tags,
			}
			resources = append(resources, mgnTemplate)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNLaunchConfigurationTemplate struct {
	svc      *mgn.Client                        `description:"-"`
	template *types.LaunchConfigurationTemplate `description:"-"`

	// Exposed properties
	LaunchConfigurationTemplateID       *string           `description:"The unique identifier of the launch configuration template"`
	Arn                                 *string           `description:"The ARN of the launch configuration template"`
	Ec2LaunchTemplateID                 *string           `description:"The ID of the associated EC2 launch template"`
	LaunchDisposition                   string            `description:"The launch disposition (STOPPED, STARTED)"`
	TargetInstanceTypeRightSizingMethod string            `description:"The method for right-sizing the target instance type"`
	CopyPrivateIP                       *bool             `description:"Whether to copy the private IP address"`
	CopyTags                            *bool             `description:"Whether to copy tags to the launched instance"`
	EnableMapAutoTagging                *bool             `description:"Whether to enable automatic tagging"`
	Tags                                map[string]string `description:"The tags associated with the template"`
}

func (r *MGNLaunchConfigurationTemplate) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteLaunchConfigurationTemplate(ctx, &mgn.DeleteLaunchConfigurationTemplateInput{
		LaunchConfigurationTemplateID: r.template.LaunchConfigurationTemplateID,
	})

	return err
}

func (r *MGNLaunchConfigurationTemplate) Properties() libtypes.Properties {
	props := libtypes.NewPropertiesFromStruct(r)
	props.Set("CopyPrivateIp", r.CopyPrivateIP)
	return props
}

func (r *MGNLaunchConfigurationTemplate) String() string {
	return *r.LaunchConfigurationTemplateID
}
