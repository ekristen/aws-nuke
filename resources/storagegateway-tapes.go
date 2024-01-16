package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const StorageGatewayTapeResource = "StorageGatewayTape"

func init() {
	resource.Register(resource.Registration{
		Name:   StorageGatewayTapeResource,
		Scope:  nuke.Account,
		Lister: &StorageGatewayTapeLister{},
	})
}

type StorageGatewayTapeLister struct{}

func (l *StorageGatewayTapeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := storagegateway.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &storagegateway.ListTapesInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListTapes(params)
		if err != nil {
			return nil, err
		}

		for _, tapeInfo := range output.TapeInfos {
			resources = append(resources, &StorageGatewayTape{
				svc:        svc,
				tapeARN:    tapeInfo.TapeARN,
				gatewayARN: tapeInfo.GatewayARN,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type StorageGatewayTape struct {
	svc        *storagegateway.StorageGateway
	tapeARN    *string
	gatewayARN *string
}

func (f *StorageGatewayTape) Remove(_ context.Context) error {
	_, err := f.svc.DeleteTape(&storagegateway.DeleteTapeInput{
		TapeARN:    f.tapeARN,
		GatewayARN: f.gatewayARN,
	})

	return err
}

func (f *StorageGatewayTape) String() string {
	return *f.tapeARN
}
