package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RedshiftSnapshotResource = "RedshiftSnapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:   RedshiftSnapshotResource,
		Scope:  nuke.Account,
		Lister: &RedshiftSnapshotLister{},
	})
}

type RedshiftSnapshotLister struct{}

func (l *RedshiftSnapshotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshift.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshift.DescribeClusterSnapshotsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeClusterSnapshots(params)
		if err != nil {
			return nil, err
		}

		for _, snapshot := range output.Snapshots {
			resources = append(resources, &RedshiftSnapshot{
				svc:      svc,
				snapshot: snapshot,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type RedshiftSnapshot struct {
	svc      *redshift.Redshift
	snapshot *redshift.Snapshot
}

func (f *RedshiftSnapshot) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreatedTime", f.snapshot.SnapshotCreateTime)

	for _, tag := range f.snapshot.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (f *RedshiftSnapshot) Remove(_ context.Context) error {
	_, err := f.svc.DeleteClusterSnapshot(&redshift.DeleteClusterSnapshotInput{
		SnapshotIdentifier: f.snapshot.SnapshotIdentifier,
	})

	return err
}

func (f *RedshiftSnapshot) String() string {
	return *f.snapshot.SnapshotIdentifier
}
