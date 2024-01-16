package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RedshiftSubnetGroupResource = "RedshiftSubnetGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   RedshiftSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &RedshiftSubnetGroupLister{},
	})
}

type RedshiftSubnetGroupLister struct{}

func (l *RedshiftSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshift.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshift.DescribeClusterSubnetGroupsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeClusterSubnetGroups(params)
		if err != nil {
			return nil, err
		}

		for _, subnetGroup := range output.ClusterSubnetGroups {
			resources = append(resources, &RedshiftSubnetGroup{
				svc:                    svc,
				clusterSubnetGroupName: subnetGroup.ClusterSubnetGroupName,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type RedshiftSubnetGroup struct {
	svc                    *redshift.Redshift
	clusterSubnetGroupName *string
}

func (f *RedshiftSubnetGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteClusterSubnetGroup(&redshift.DeleteClusterSubnetGroupInput{
		ClusterSubnetGroupName: f.clusterSubnetGroupName,
	})

	return err
}

func (f *RedshiftSubnetGroup) String() string {
	return *f.clusterSubnetGroupName
}
