package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VolumeResource = "EC2Volume"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2VolumeResource,
		Scope:  nuke.Account,
		Lister: &EC2VolumeLister{},
	})
}

type EC2Volume struct {
	svc    *ec2.EC2
	volume *ec2.Volume
}

type EC2VolumeLister struct{}

func (l *EC2VolumeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeVolumes(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.Volumes {
		resources = append(resources, &EC2Volume{
			svc:    svc,
			volume: out,
		})
	}

	return resources, nil
}

func (e *EC2Volume) Remove(_ context.Context) error {
	_, err := e.svc.DeleteVolume(&ec2.DeleteVolumeInput{
		VolumeId: e.volume.VolumeId,
	})
	return err
}

func (e *EC2Volume) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("State", e.volume.State)
	properties.Set("CreateTime", e.volume.CreateTime.Format(time.RFC3339))
	for _, tagValue := range e.volume.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *EC2Volume) String() string {
	return *e.volume.VolumeId
}
