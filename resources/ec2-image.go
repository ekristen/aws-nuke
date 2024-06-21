package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2ImageResource = "EC2Image"

const IncludeDeprecatedSetting = "IncludeDeprecated"
const IncludeDisabledSetting = "IncludeDisabled"
const DisableDeregistrationProtectionSetting = "DisableDeregistrationProtection"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2ImageResource,
		Scope:  nuke.Account,
		Lister: &EC2ImageLister{},
		Settings: []string{
			DisableDeregistrationProtectionSetting,
			IncludeDeprecatedSetting,
			IncludeDisabledSetting,
		},
	})
}

type EC2ImageLister struct{}

func (l *EC2ImageLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	params := &ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String("self"),
		},
		IncludeDeprecated: ptr.Bool(true),
		IncludeDisabled:   ptr.Bool(true),
	}
	resp, err := svc.DescribeImages(params)
	if err != nil {
		return nil, err
	}

	for _, out := range resp.Images {
		visibility := "Private"
		if ptr.ToBool(out.Public) {
			visibility = "Public"
		}

		resources = append(resources, &EC2Image{
			svc:                      svc,
			id:                       out.ImageId,
			name:                     out.Name,
			tags:                     out.Tags,
			state:                    out.State,
			visibility:               ptr.String(visibility),
			creationDate:             out.CreationDate,
			deprecated:               ptr.Bool(out.DeprecationTime != nil),
			deprecatedTime:           out.DeprecationTime,
			deregistrationProtection: out.DeregistrationProtection,
		})
	}

	return resources, nil
}

type EC2Image struct {
	svc      *ec2.EC2
	settings *libsettings.Setting

	id                       *string
	name                     *string
	tags                     []*ec2.Tag
	state                    *string
	visibility               *string
	deprecated               *bool
	deprecatedTime           *string
	creationDate             *string
	deregistrationProtection *string
}

func (e *EC2Image) Filter() error {
	if *e.state == "pending" {
		return fmt.Errorf("ineligible state for removal")
	}

	if strings.HasPrefix(*e.deregistrationProtection, "disabled after") {
		return fmt.Errorf("would remove after %s due to deregistration protection cooldown",
			strings.ReplaceAll(*e.deregistrationProtection, "disabled after ", ""))
	}

	if *e.deregistrationProtection != ec2.ImageStateDisabled {
		if e.settings.Get(DisableDeregistrationProtectionSetting) == nil ||
			(e.settings.Get(DisableDeregistrationProtectionSetting) != nil &&
				!e.settings.Get(DisableDeregistrationProtectionSetting).(bool)) {
			return fmt.Errorf("deregistration protection is enabled")
		}
	}

	if !e.settings.Get(IncludeDeprecatedSetting).(bool) && e.deprecated != nil && *e.deprecated {
		return fmt.Errorf("excluded by %s setting being false", IncludeDeprecatedSetting)
	}

	if !e.settings.Get(IncludeDisabledSetting).(bool) && e.state != nil && *e.state == ec2.ImageStateDisabled {
		return fmt.Errorf("excluded by %s setting being false", IncludeDisabledSetting)
	}

	return nil
}

func (e *EC2Image) Remove(_ context.Context) error {
	if err := e.removeDeregistrationProtection(); err != nil {
		return err
	}

	if *e.deregistrationProtection == "enabled-with-cooldown" {
		return nil
	}

	_, err := e.svc.DeregisterImage(&ec2.DeregisterImageInput{
		ImageId: e.id,
	})

	return err
}

func (e *EC2Image) removeDeregistrationProtection() error {
	if *e.deregistrationProtection == ec2.ImageStateDisabled {
		return nil
	}

	if !e.settings.Get(DisableDeregistrationProtectionSetting).(bool) {
		return nil
	}

	_, err := e.svc.DisableImageDeregistrationProtection(&ec2.DisableImageDeregistrationProtectionInput{
		ImageId: e.id,
	})
	return err
}

func (e *EC2Image) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("CreationDate", e.creationDate)
	properties.Set("Name", e.name)
	properties.Set("State", e.state)
	properties.Set("Visibility", e.visibility)
	properties.Set("Deprecated", e.deprecated)
	properties.Set("DeprecatedTime", e.deprecatedTime)
	properties.Set("DeregistrationProtection", e.deregistrationProtection)

	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *EC2Image) String() string {
	return *e.id
}

func (e *EC2Image) Settings(settings *libsettings.Setting) {
	e.settings = settings
}
