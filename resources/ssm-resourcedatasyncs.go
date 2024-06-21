package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMResourceDataSyncResource = "SSMResourceDataSync"

func init() {
	registry.Register(&registry.Registration{
		Name:   SSMResourceDataSyncResource,
		Scope:  nuke.Account,
		Lister: &SSMResourceDataSyncLister{},
	})
}

type SSMResourceDataSyncLister struct{}

func (l *SSMResourceDataSyncLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ssm.ListResourceDataSyncInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListResourceDataSync(params)
		if err != nil {
			return nil, err
		}

		for _, resourceDataSyncItem := range output.ResourceDataSyncItems {
			resources = append(resources, &SSMResourceDataSync{
				svc:  svc,
				name: resourceDataSyncItem.SyncName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMResourceDataSync struct {
	svc  *ssm.SSM
	name *string
}

func (f *SSMResourceDataSync) Remove(_ context.Context) error {
	_, err := f.svc.DeleteResourceDataSync(&ssm.DeleteResourceDataSyncInput{
		SyncName: f.name,
	})

	return err
}

func (f *SSMResourceDataSync) String() string {
	return *f.name
}
