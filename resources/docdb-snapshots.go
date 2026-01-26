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

	params := &docdb.DescribeDBClusterSnapshotsInput{
		Filters: []docdbtypes.Filter{
			{
				Name:   aws.String("engine"),
				Values: []string{"docdb"},
			},
		},
	}

	paginator := docdb.NewDescribeDBClusterSnapshotsPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(page.DBClusterSnapshots); i++ {
			tagList := DocDBEmptyTags
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: page.DBClusterSnapshots[i].DBClusterSnapshotArn,
			})
			if err == nil {
				tagList = tags.TagList
			}
			resources = append(resources, &DocDBSnapshot{
				svc:                svc,
				ARN:                page.DBClusterSnapshots[i].DBClusterSnapshotArn,
				Identifier:         page.DBClusterSnapshots[i].DBClusterIdentifier,
				SnapshotType:       page.DBClusterSnapshots[i].SnapshotType,
				Status:             page.DBClusterSnapshots[i].Status,
				AvailabilityZones:  page.DBClusterSnapshots[i].AvailabilityZones,
				SnapshotCreateTime: page.DBClusterSnapshots[i].SnapshotCreateTime,
				Tags:               tagList,
			})
		}
	}
	return resources, nil
}

type DocDBSnapshot struct {
	svc *docdb.Client

	ARN                *string
	Identifier         *string
	SnapshotType       *string
	Status             *string
	AvailabilityZones  []string
	SnapshotCreateTime *time.Time
	Tags               []docdbtypes.Tag
}

func (r *DocDBSnapshot) Filter() error {
	if *r.SnapshotType == RDSAutomatedSnapshot {
		return fmt.Errorf("cannot delete automated snapshots")
	}
	return nil
}

func (r *DocDBSnapshot) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDBClusterSnapshot(ctx, &docdb.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: r.Identifier,
	})
	return err
}

func (r *DocDBSnapshot) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DocDBSnapshot) String() string {
	return *r.Identifier
}
