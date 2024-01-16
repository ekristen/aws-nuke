package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const NeptuneInstanceResource = "NeptuneInstance"

func init() {
	resource.Register(resource.Registration{
		Name:   NeptuneInstanceResource,
		Scope:  nuke.Account,
		Lister: &NeptuneInstanceLister{},
	})
}

type NeptuneInstanceLister struct{}

func (l *NeptuneInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := neptune.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &neptune.DescribeDBInstancesInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDBInstances(params)
		if err != nil {
			return nil, err
		}

		for _, dbInstance := range output.DBInstances {
			resources = append(resources, &NeptuneInstance{
				svc: svc,
				ID:  dbInstance.DBInstanceIdentifier,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type NeptuneInstance struct {
	svc *neptune.Neptune
	ID  *string
}

func (f *NeptuneInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDBInstance(&neptune.DeleteDBInstanceInput{
		DBInstanceIdentifier: f.ID,
		SkipFinalSnapshot:    aws.Bool(true),
	})

	return err
}

func (f *NeptuneInstance) String() string {
	return *f.ID
}
