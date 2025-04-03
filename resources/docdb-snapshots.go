package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	docdbtypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DocDBSnapshotResource = "DocDBSnapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBSnapshotResource,
		Scope:    nuke.Account,
		Resource: &DocDBSnapshot{},
		Lister:   &DocDBSnapshotLister{},
	})
}

type DocDBSnapshotLister struct{}

func (l *DocDBSnapshotLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdb.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := docdb.NewDescribeDBClusterSnapshotsPaginator(svc, &docdb.DescribeDBClusterSnapshotsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(page.DBClusterSnapshots); i++ {
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: page.DBClusterSnapshots[i].DBClusterSnapshotArn,
			})
			if err != nil {
				continue
			}
			resources = append(resources, &DocDBSnapshot{
				svc:      svc,
				snapshot: page.DBClusterSnapshots[i],
				tags:     tags.TagList,
			})
		}
	}
	return resources, nil
}

type DocDBSnapshot struct {
	svc      *docdb.Client
	snapshot docdbtypes.DBClusterSnapshot
	tags     []docdbtypes.Tag
}

const DocDBAutomatedSnapshot = "automated"

func (r *DocDBSnapshot) Filter() error {
	if *r.snapshot.SnapshotType == DocDBAutomatedSnapshot {
		return fmt.Errorf("cannot delete automated snapshots")
	}
	return nil
}

func (r *DocDBSnapshot) Remove(ctx context.Context) error {
	if r.snapshot.DBClusterIdentifier == nil {
		return nil
	}
	_, err := r.svc.DeleteDBClusterSnapshot(ctx, &docdb.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: aws.String(*r.snapshot.DBClusterIdentifier),
	})
	return err
}

func (r *DocDBSnapshot) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ARN", r.snapshot.DBClusterSnapshotArn)
	properties.Set("Identifier", r.snapshot.DBClusterSnapshotIdentifier)
	properties.Set("SnapshotType", r.snapshot.SnapshotType)
	properties.Set("Status", r.snapshot.Status)
	properties.Set("AvailabilityZones", r.snapshot.AvailabilityZones)

	if r.snapshot.SnapshotCreateTime != nil {
		properties.Set("SnapshotCreateTime", r.snapshot.SnapshotCreateTime.Format(time.RFC3339))
	}

	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
