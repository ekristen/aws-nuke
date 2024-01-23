package resources

import (
	"context"

	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2SnapshotResource = "EC2Snapshot"

func init() {
	resource.Register(&resource.Registration{
		Name:   EC2SnapshotResource,
		Scope:  nuke.Account,
		Lister: &EC2SnapshotLister{},
		DependsOn: []string{
			EC2ImageResource,
		},
	})
}

type EC2SnapshotLister struct{}

func (l *EC2SnapshotLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	params := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{
			aws.String("self"),
		},
	}
	resp, err := svc.DescribeSnapshots(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.Snapshots {
		resources = append(resources, &EC2Snapshot{
			svc:       svc,
			id:        *out.SnapshotId,
			startTime: out.StartTime,
			tags:      out.Tags,
		})
	}

	return resources, nil
}

type EC2Snapshot struct {
	svc       *ec2.EC2
	id        string
	startTime *time.Time
	tags      []*ec2.Tag
}

func (e *EC2Snapshot) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("StartTime", e.startTime.Format(time.RFC3339))

	for _, tagValue := range e.tags {
		properties.Set(fmt.Sprintf("tag:%v", *tagValue.Key), tagValue.Value)
	}
	return properties
}

func (e *EC2Snapshot) Remove(_ context.Context) error {
	_, err := e.svc.DeleteSnapshot(&ec2.DeleteSnapshotInput{
		SnapshotId: &e.id,
	})
	return err
}

func (e *EC2Snapshot) String() string {
	return e.id
}
