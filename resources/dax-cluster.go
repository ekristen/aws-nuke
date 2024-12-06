package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DAXClusterResource = "DAXCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     DAXClusterResource,
		Scope:    nuke.Account,
		Resource: &DAXCluster{},
		Lister:   &DAXClusterLister{},
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
				svc:  svc,
				Name: cluster.ClusterName,
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
	svc  *dax.DAX
	Name *string
}

func (r *DAXCluster) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCluster(&dax.DeleteClusterInput{
		ClusterName: r.Name,
	})

	return err
}

func (r *DAXCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DAXCluster) String() string {
	return *r.Name
}
