package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2SnapshotResource = "EC2Snapshot"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2SnapshotResource,
		Scope:    nuke.Account,
		Resource: &EC2Snapshot{},
		Lister:   &EC2SnapshotLister{},
	})
}

type EC2SnapshotLister struct{}

func (l *EC2SnapshotLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	params := &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeSnapshots(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.Snapshots {
			snapshot := &resp.Snapshots[i]
			resources = append(resources, &EC2Snapshot{
				svc:                 svc,
				SnapshotID:          snapshot.SnapshotId,
				Description:         snapshot.Description,
				VolumeID:            snapshot.VolumeId,
				VolumeSize:          snapshot.VolumeSize,
				State:               &snapshot.State,
				StateMessage:        snapshot.StateMessage,
				StartTime:           snapshot.StartTime,
				Progress:            snapshot.Progress,
				OwnerID:             snapshot.OwnerId,
				OwnerAlias:          snapshot.OwnerAlias,
				Encrypted:           snapshot.Encrypted,
				KmsKeyID:            snapshot.KmsKeyId,
				DataEncryptionKeyID: snapshot.DataEncryptionKeyId,
				StorageTier:         &snapshot.StorageTier,
				RestoreExpiryTime:   snapshot.RestoreExpiryTime,
				Tags:                &snapshot.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2Snapshot struct {
	svc                 *ec2.Client
	SnapshotID          *string                 `description:"The ID of the snapshot"`
	Description         *string                 `description:"The description for the snapshot"`
	VolumeID            *string                 `description:"The ID of the volume that was used to create the snapshot"`
	VolumeSize          *int32                  `description:"The size of the volume in GiB"`
	State               *ec2types.SnapshotState `description:"The snapshot state"`
	StateMessage        *string                 `description:"Encrypted Amazon EBS snapshots are copied asynchronously"`
	StartTime           *time.Time              `description:"The time stamp when the snapshot was initiated"`
	Progress            *string                 `description:"The progress of the snapshot as a percentage"`
	OwnerID             *string                 `description:"The AWS account ID of the EBS snapshot owner"`
	OwnerAlias          *string                 `description:"The AWS owner alias"`
	Encrypted           *bool                   `description:"Indicates whether the snapshot is encrypted"`
	KmsKeyID            *string                 `description:"The Amazon Resource Name (ARN) of the AWS KMS key used for encryption"`
	DataEncryptionKeyID *string                 `description:"The data encryption key identifier for the snapshot"`
	StorageTier         *ec2types.StorageTier   `description:"The storage tier in which the snapshot is stored"`
	RestoreExpiryTime   *time.Time              `description:"Only for archived snapshots that are temporarily restored"`
	Tags                *[]ec2types.Tag         `description:"The tags associated with the snapshot"`
}

func (r *EC2Snapshot) Remove(ctx context.Context) error {
	params := &ec2.DeleteSnapshotInput{
		SnapshotId: r.SnapshotID,
	}

	_, err := r.svc.DeleteSnapshot(ctx, params)
	return err
}

func (r *EC2Snapshot) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2Snapshot) String() string {
	return *r.SnapshotID
}
