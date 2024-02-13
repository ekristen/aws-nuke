package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const StorageGatewayVolumeResource = "StorageGatewayVolume"

func init() {
	registry.Register(&registry.Registration{
		Name:   StorageGatewayVolumeResource,
		Scope:  nuke.Account,
		Lister: &StorageGatewayVolumeLister{},
	})
}

type StorageGatewayVolumeLister struct{}

func (l *StorageGatewayVolumeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := storagegateway.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &storagegateway.ListVolumesInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListVolumes(params)
		if err != nil {
			return nil, err
		}

		for _, volumeInfo := range output.VolumeInfos {
			resources = append(resources, &StorageGatewayVolume{
				svc: svc,
				ARN: volumeInfo.VolumeARN,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type StorageGatewayVolume struct {
	svc *storagegateway.StorageGateway
	ARN *string
}

func (f *StorageGatewayVolume) Remove(_ context.Context) error {

	_, err := f.svc.DeleteVolume(&storagegateway.DeleteVolumeInput{
		VolumeARN: f.ARN,
	})

	return err
}

func (f *StorageGatewayVolume) String() string {
	return *f.ARN
}
