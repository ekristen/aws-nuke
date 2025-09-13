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

const EC2VolumeResource = "EC2Volume"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VolumeResource,
		Scope:    nuke.Account,
		Resource: &EC2Volume{},
		Lister:   &EC2VolumeLister{},
	})
}

type EC2VolumeLister struct{}

func (l *EC2VolumeLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	params := &ec2.DescribeVolumesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeVolumes(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.Volumes {
			volume := &resp.Volumes[i]
			resources = append(resources, &EC2Volume{
				svc:                svc,
				VolumeID:           volume.VolumeId,
				VolumeType:         &volume.VolumeType,
				State:              &volume.State,
				Size:               volume.Size,
				AvailabilityZone:   volume.AvailabilityZone,
				CreateTime:         volume.CreateTime,
				Encrypted:          volume.Encrypted,
				KmsKeyID:           volume.KmsKeyId,
				Iops:               volume.Iops,
				Throughput:         volume.Throughput,
				MultiAttachEnabled: volume.MultiAttachEnabled,
				Tags:               &volume.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2Volume struct {
	svc                *ec2.Client
	VolumeID           *string               `description:"The ID of the EBS volume"`
	VolumeType         *ec2types.VolumeType  `description:"The volume type (gp2, gp3, io1, io2, st1, sc1, standard)"`
	State              *ec2types.VolumeState `description:"The state of the volume (creating, available, in-use, deleting, deleted, error)"`
	Size               *int32                `description:"The size of the volume in GiB"`
	AvailabilityZone   *string               `description:"The Availability Zone in which the volume was created"`
	CreateTime         *time.Time            `description:"The time stamp when volume creation was initiated"`
	Encrypted          *bool                 `description:"Indicates whether the volume is encrypted"`
	KmsKeyID           *string               `description:"The Amazon Resource Name (ARN) of the AWS KMS key used for encryption"`
	Iops               *int32                `description:"The number of I/O operations per second (IOPS)"`
	Throughput         *int32                `description:"The throughput that the volume supports in MiB/s"`
	MultiAttachEnabled *bool                 `description:"Indicates whether Amazon EBS Multi-Attach is enabled"`
	Tags               *[]ec2types.Tag       `description:"The tags associated with the EBS volume"`
}

func (r *EC2Volume) Remove(ctx context.Context) error {
	params := &ec2.DeleteVolumeInput{
		VolumeId: r.VolumeID,
	}

	_, err := r.svc.DeleteVolume(ctx, params)
	return err
}

func (r *EC2Volume) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2Volume) String() string {
	return *r.VolumeID
}
