package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NeptuneClusterResource = "NeptuneCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:   NeptuneClusterResource,
		Scope:  nuke.Account,
		Lister: &NeptuneClusterLister{},
	})
}

type NeptuneClusterLister struct{}

func (l *NeptuneClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := neptune.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &neptune.DescribeDBClustersInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDBClusters(params)
		if err != nil {
			return nil, err
		}

		for _, dbCluster := range output.DBClusters {
			resources = append(resources, &NeptuneCluster{
				svc: svc,
				ID:  dbCluster.DBClusterIdentifier,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type NeptuneCluster struct {
	svc *neptune.Neptune
	ID  *string
}

func (f *NeptuneCluster) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDBCluster(&neptune.DeleteDBClusterInput{
		DBClusterIdentifier: f.ID,
		SkipFinalSnapshot:   aws.Bool(true),
	})

	return err
}

func (f *NeptuneCluster) String() string {
	return *f.ID
}
