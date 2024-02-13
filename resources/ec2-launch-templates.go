package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2LaunchTemplateResource = "EC2LaunchTemplate"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2LaunchTemplateResource,
		Scope:  nuke.Account,
		Lister: &EC2LaunchTemplateLister{},
	})
}

type EC2LaunchTemplateLister struct{}

func (l *EC2LaunchTemplateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeLaunchTemplates(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, template := range resp.LaunchTemplates {
		resources = append(resources, &EC2LaunchTemplate{
			svc:  svc,
			name: template.LaunchTemplateName,
			tag:  template.Tags,
		})
	}
	return resources, nil
}

type EC2LaunchTemplate struct {
	svc  *ec2.EC2
	name *string
	tag  []*ec2.Tag
}

func (template *EC2LaunchTemplate) Remove(_ context.Context) error {
	_, err := template.svc.DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{
		LaunchTemplateName: template.name,
	})
	return err
}

func (template *EC2LaunchTemplate) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range template.tag {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", template.name)
	return properties
}

func (template *EC2LaunchTemplate) String() string {
	return ptr.ToString(template.name)
}
