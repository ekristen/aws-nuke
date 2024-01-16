package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const StorageGatewayGatewayResource = "StorageGatewayGateway"

func init() {
	resource.Register(resource.Registration{
		Name:   StorageGatewayGatewayResource,
		Scope:  nuke.Account,
		Lister: &StorageGatewayGatewayLister{},
	})
}

type StorageGatewayGatewayLister struct{}

func (l *StorageGatewayGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := storagegateway.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &storagegateway.ListGatewaysInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListGateways(params)
		if err != nil {
			return nil, err
		}

		for _, gateway := range output.Gateways {
			resources = append(resources, &StorageGatewayGateway{
				svc: svc,
				ARN: gateway.GatewayARN,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type StorageGatewayGateway struct {
	svc *storagegateway.StorageGateway
	ARN *string
}

func (f *StorageGatewayGateway) Remove(_ context.Context) error {
	_, err := f.svc.DeleteGateway(&storagegateway.DeleteGatewayInput{
		GatewayARN: f.ARN,
	})

	return err
}

func (f *StorageGatewayGateway) String() string {
	return *f.ARN
}
