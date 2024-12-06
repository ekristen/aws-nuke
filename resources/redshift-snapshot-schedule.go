package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RedshiftSnapshotScheduleResource = "RedshiftSnapshotSchedule"

func init() {
	registry.Register(&registry.Registration{
		Name:     RedshiftSnapshotScheduleResource,
		Scope:    nuke.Account,
		Resource: &RedshiftSnapshotSchedule{},
		Lister:   &RedshiftSnapshotScheduleLister{},
	})
}

type RedshiftSnapshotScheduleLister struct{}

func (l *RedshiftSnapshotScheduleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshift.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshift.DescribeSnapshotSchedulesInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeSnapshotSchedules(params)
		if err != nil {
			return nil, err
		}

		for _, snapshotSchedule := range output.SnapshotSchedules {
			resources = append(resources, &RedshiftSnapshotSchedule{
				svc:                svc,
				ID:                 snapshotSchedule.ScheduleIdentifier,
				Tags:               snapshotSchedule.Tags,
				associatedClusters: snapshotSchedule.AssociatedClusters,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type RedshiftSnapshotSchedule struct {
	svc                *redshift.Redshift
	ID                 *string
	Tags               []*redshift.Tag
	associatedClusters []*redshift.ClusterAssociatedToSchedule
}

func (r *RedshiftSnapshotSchedule) Properties() types.Properties {
	associatedClusters := make([]string, len(r.associatedClusters))
	for i, cluster := range r.associatedClusters {
		associatedClusters[i] = *cluster.ClusterIdentifier
	}
	properties := types.NewProperties()
	properties.Set("ID", r.ID)
	properties.Set("AssociatedClusters", associatedClusters)
	for _, tag := range r.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	return properties
}

func (r *RedshiftSnapshotSchedule) Remove(_ context.Context) error {
	for _, associatedCluster := range r.associatedClusters {
		_, disassociateErr := r.svc.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
			ScheduleIdentifier:   r.ID,
			ClusterIdentifier:    associatedCluster.ClusterIdentifier,
			DisassociateSchedule: aws.Bool(true),
		})

		if disassociateErr != nil {
			return disassociateErr
		}
	}

	_, err := r.svc.DeleteSnapshotSchedule(&redshift.DeleteSnapshotScheduleInput{
		ScheduleIdentifier: r.ID,
	})

	return err
}

func (r *RedshiftSnapshotSchedule) String() string {
	return *r.ID
}
