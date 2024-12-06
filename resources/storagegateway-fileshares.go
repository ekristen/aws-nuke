package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const StorageGatewayFileShareResource = "StorageGatewayFileShare"

func init() {
	registry.Register(&registry.Registration{
		Name:     StorageGatewayFileShareResource,
		Scope:    nuke.Account,
		Resource: &StorageGatewayFileShare{},
		Lister:   &StorageGatewayFileShareLister{},
	})
}

type StorageGatewayFileShareLister struct{}

func (l *StorageGatewayFileShareLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := storagegateway.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &storagegateway.ListFileSharesInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListFileShares(params)
		if err != nil {
			return nil, err
		}

		for _, fileShareInfo := range output.FileShareInfoList {
			resources = append(resources, &StorageGatewayFileShare{
				svc: svc,
				ARN: fileShareInfo.FileShareARN,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type StorageGatewayFileShare struct {
	svc *storagegateway.StorageGateway
	ARN *string
}

func (f *StorageGatewayFileShare) Remove(_ context.Context) error {
	_, err := f.svc.DeleteFileShare(&storagegateway.DeleteFileShareInput{
		FileShareARN: f.ARN,
		ForceDelete:  aws.Bool(true),
	})

	return err
}

func (f *StorageGatewayFileShare) String() string {
	return *f.ARN
}
