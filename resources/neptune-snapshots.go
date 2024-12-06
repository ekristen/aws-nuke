package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NeptuneSnapshotResource = "NeptuneSnapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:     NeptuneSnapshotResource,
		Scope:    nuke.Account,
		Resource: &NeptuneSnapshot{},
		Lister:   &NeptuneSnapshotLister{},
		DeprecatedAliases: []string{
			"NetpuneSnapshot",
		},
	})
}

type NeptuneSnapshotLister struct{}

func (l *NeptuneSnapshotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := neptune.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &neptune.DescribeDBClusterSnapshotsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDBClusterSnapshots(params)
		if err != nil {
			return nil, err
		}

		for _, dbClusterSnapshot := range output.DBClusterSnapshots {
			resources = append(resources, &NeptuneSnapshot{
				svc: svc,
				ID:  dbClusterSnapshot.DBClusterSnapshotIdentifier,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type NeptuneSnapshot struct {
	svc *neptune.Neptune
	ID  *string
}

func (f *NeptuneSnapshot) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDBClusterSnapshot(&neptune.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: f.ID,
	})

	return err
}

func (f *NeptuneSnapshot) String() string {
	return *f.ID
}
