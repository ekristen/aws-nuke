package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lightsail"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const LightsailDiskResource = "LightsailDisk"

func init() {
	resource.Register(resource.Registration{
		Name:   LightsailDiskResource,
		Scope:  nuke.Account,
		Lister: &LightsailDiskLister{},
	})
}

type LightsailDiskLister struct{}

func (l *LightsailDiskLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lightsail.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lightsail.GetDisksInput{}

	for {
		output, err := svc.GetDisks(params)
		if err != nil {
			return nil, err
		}

		for _, disk := range output.Disks {
			resources = append(resources, &LightsailDisk{
				svc:      svc,
				diskName: disk.Name,
			})
		}

		if output.NextPageToken == nil {
			break
		}

		params.PageToken = output.NextPageToken
	}

	return resources, nil
}

type LightsailDisk struct {
	svc      *lightsail.Lightsail
	diskName *string
}

func (f *LightsailDisk) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDisk(&lightsail.DeleteDiskInput{
		DiskName: f.diskName,
	})

	return err
}

func (f *LightsailDisk) String() string {
	return *f.diskName
}
