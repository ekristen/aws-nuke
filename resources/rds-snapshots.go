package resources

import (
	"context"

	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSSnapshotResource = "RDSSnapshot"

func init() {
	resource.Register(&resource.Registration{
		Name:   RDSSnapshotResource,
		Scope:  nuke.Account,
		Lister: &RDSSnapshotLister{},
	})
}

type RDSSnapshotLister struct{}

func (l *RDSSnapshotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBSnapshotsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeDBSnapshots(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, snapshot := range resp.DBSnapshots {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: snapshot.DBSnapshotArn,
		})
		if err != nil {
			return nil, err
		}

		resources = append(resources, &RDSSnapshot{
			svc:      svc,
			snapshot: snapshot,
			tags:     tags.TagList,
		})

	}

	return resources, nil
}

type RDSSnapshot struct {
	svc      *rds.RDS
	snapshot *rds.DBSnapshot
	tags     []*rds.Tag
}

func (i *RDSSnapshot) Filter() error {
	if *i.snapshot.SnapshotType == "automated" {
		return fmt.Errorf("cannot delete automated snapshots")
	}
	return nil
}

func (i *RDSSnapshot) Remove(_ context.Context) error {
	if i.snapshot.DBSnapshotIdentifier == nil {
		// Sanity check to make sure the delete request does not skip the
		// identifier.
		return nil
	}

	params := &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: i.snapshot.DBSnapshotIdentifier,
	}

	_, err := i.svc.DeleteDBSnapshot(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSSnapshot) String() string {
	return *i.snapshot.DBSnapshotIdentifier
}

func (i *RDSSnapshot) Properties() types.Properties {
	properties := types.NewProperties().
		Set("ARN", i.snapshot.DBSnapshotArn).
		Set("Identifier", i.snapshot.DBSnapshotIdentifier).
		Set("SnapshotType", i.snapshot.SnapshotType).
		Set("Status", i.snapshot.Status).
		Set("AvailabilityZone", i.snapshot.AvailabilityZone).
		Set("CreatedTime", i.snapshot.SnapshotCreateTime.Format(time.RFC3339))

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
