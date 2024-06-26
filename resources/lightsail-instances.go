package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/lightsail"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LightsailInstanceResource = "LightsailInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:   LightsailInstanceResource,
		Scope:  nuke.Account,
		Lister: &LightsailInstanceLister{},
	})
}

type LightsailInstanceLister struct{}

func (l *LightsailInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lightsail.GetInstancesInput{}

	for {
		output, err := svc.GetInstances(params)
		if err != nil {
			return nil, err
		}

		for _, instance := range output.Instances {
			resources = append(resources, &LightsailInstance{
				svc:          svc,
				instanceName: instance.Name,
				tags:         instance.Tags,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailInstance struct {
	svc          *lightsail.Lightsail
	instanceName *string
	tags         []*lightsail.Tag

	settings *libsettings.Setting
}

func (f *LightsailInstance) Settings(setting *libsettings.Setting) {
	f.settings = setting
}

func (f *LightsailInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeleteInstance(&lightsail.DeleteInstanceInput{
		InstanceName:      f.instanceName,
		ForceDeleteAddOns: ptr.Bool(f.settings.GetBool("ForceDeleteAddOns")),
	})

	return err
}

func (f *LightsailInstance) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.Set("Name", f.instanceName)
	return properties
}

func (f *LightsailInstance) String() string {
	return *f.instanceName
}
