package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTOTAUpdateResource = "IoTOTAUpdate"

func init() {
	resource.Register(resource.Registration{
		Name:   IoTOTAUpdateResource,
		Scope:  nuke.Account,
		Lister: &IoTOTAUpdateLister{},
	})
}

type IoTOTAUpdateLister struct{}

func (l *IoTOTAUpdateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListOTAUpdatesInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListOTAUpdates(params)
		if err != nil {
			return nil, err
		}

		for _, otaUpdate := range output.OtaUpdates {
			resources = append(resources, &IoTOTAUpdate{
				svc: svc,
				ID:  otaUpdate.OtaUpdateId,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type IoTOTAUpdate struct {
	svc *iot.IoT
	ID  *string
}

func (f *IoTOTAUpdate) Remove(_ context.Context) error {
	_, err := f.svc.DeleteOTAUpdate(&iot.DeleteOTAUpdateInput{
		OtaUpdateId: f.ID,
	})

	return err
}

func (f *IoTOTAUpdate) String() string {
	return *f.ID
}
