package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DAXClusterResource = "DAXCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:   DAXClusterResource,
		Scope:  nuke.Account,
		Lister: &DAXClusterLister{},
		DependsOn: []string{
			DAXParameterGroupResource,
			DAXSubnetGroupResource,
		},
	})
}

type DAXClusterLister struct{}

func (l *DAXClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := dax.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &dax.DescribeClustersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range output.Clusters {
			resources = append(resources, &DAXCluster{
				svc:         svc,
				clusterName: cluster.ClusterName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type DAXCluster struct {
	svc         *dax.DAX
	clusterName *string
}

func (f *DAXCluster) Remove(_ context.Context) error {
	_, err := f.svc.DeleteCluster(&dax.DeleteClusterInput{
		ClusterName: f.clusterName,
	})

	return err
}

func (f *DAXCluster) String() string {
	return *f.clusterName
}
