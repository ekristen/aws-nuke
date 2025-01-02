package resources

import (
	"context"
	"github.com/ekristen/libnuke/pkg/types"
	"time"

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
				svc:        svc,
				ID:         dbClusterSnapshot.DBClusterSnapshotIdentifier,
				Status:     dbClusterSnapshot.Status,
				CreateTime: dbClusterSnapshot.SnapshotCreateTime,
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
	svc        *neptune.Neptune
	ID         *string
	Status     *string
	CreateTime *time.Time
}

func (r *NeptuneSnapshot) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDBClusterSnapshot(&neptune.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: r.ID,
	})

	return err
}

func (r *NeptuneSnapshot) String() string {
	return *r.ID
}

func (r *NeptuneSnapshot) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
