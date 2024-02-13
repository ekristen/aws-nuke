package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSClusterSnapshotResource = "RDSClusterSnapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:   RDSClusterSnapshotResource,
		Scope:  nuke.Account,
		Lister: &RDSClusterSnapshotLister{},
	})
}

type RDSClusterSnapshotLister struct{}

func (l *RDSClusterSnapshotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBClusterSnapshotsInput{MaxRecords: aws.Int64(100)}

	resp, err := svc.DescribeDBClusterSnapshots(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, snapshot := range resp.DBClusterSnapshots {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: snapshot.DBClusterSnapshotArn,
		})
		if err != nil {
			return nil, err
		}

		resources = append(resources, &RDSClusterSnapshot{
			svc:      svc,
			snapshot: snapshot,
			tags:     tags.TagList,
		})

	}

	return resources, nil
}

type RDSClusterSnapshot struct {
	svc      *rds.RDS
	snapshot *rds.DBClusterSnapshot
	tags     []*rds.Tag
}

func (i *RDSClusterSnapshot) Filter() error {
	if *i.snapshot.SnapshotType == "automated" {
		return fmt.Errorf("cannot delete automated snapshots")
	}
	return nil
}

func (i *RDSClusterSnapshot) Remove(_ context.Context) error {
	if i.snapshot.DBClusterSnapshotIdentifier == nil {
		// Sanity check to make sure the delete request does not skip the
		// identifier.
		return nil
	}

	params := &rds.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: i.snapshot.DBClusterSnapshotIdentifier,
	}

	_, err := i.svc.DeleteDBClusterSnapshot(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSClusterSnapshot) String() string {
	return *i.snapshot.DBClusterSnapshotIdentifier
}

func (i *RDSClusterSnapshot) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", i.snapshot.DBClusterSnapshotArn)
	properties.Set("Identifier", i.snapshot.DBClusterSnapshotIdentifier)
	properties.Set("SnapshotType", i.snapshot.SnapshotType)
	properties.Set("Status", i.snapshot.Status)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
