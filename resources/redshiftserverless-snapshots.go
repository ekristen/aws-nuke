package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type RedshiftServerlessSnapshot struct {
	svc      *redshiftserverless.RedshiftServerless
	snapshot *redshiftserverless.Snapshot
}

const RedshiftServerlessSnapshotResource = "RedshiftServerlessSnapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:   RedshiftServerlessSnapshotResource,
		Scope:  nuke.Account,
		Lister: &RedshiftServerlessSnapshotLister{},
	})
}

type RedshiftServerlessSnapshotLister struct{}

func (l *RedshiftServerlessSnapshotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshiftserverless.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshiftserverless.ListSnapshotsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListSnapshots(params)
		if err != nil {
			return nil, err
		}

		for _, snapshot := range output.Snapshots {
			resources = append(resources, &RedshiftServerlessSnapshot{
				svc:      svc,
				snapshot: snapshot,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (s *RedshiftServerlessSnapshot) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreateTime", s.snapshot.SnapshotCreateTime).
		Set("Namespace", s.snapshot.NamespaceName).
		Set("SnapshotName", s.snapshot.SnapshotName)

	return properties
}

func (s *RedshiftServerlessSnapshot) Remove(_ context.Context) error {
	_, err := s.svc.DeleteSnapshot(&redshiftserverless.DeleteSnapshotInput{
		SnapshotName: s.snapshot.SnapshotName,
	})

	return err
}

func (s *RedshiftServerlessSnapshot) String() string {
	return ptr.ToString(s.snapshot.SnapshotName)
}
