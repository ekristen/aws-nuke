package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/docdbelastic"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const DocDBElasticClusterResource = "DocDBElasticCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBElasticClusterResource,
		Scope:    nuke.Account,
		Resource: &DocDBElasticCluster{},
		Lister:   &DocDBElasticClusterLister{},
	})
}

type DocDBElasticClusterLister struct{}

func (l *DocDBElasticClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdbelastic.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := docdbelastic.NewListClustersPaginator(svc, &docdbelastic.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, cluster := range page.Clusters {
			resources = append(resources, &DocDBElasticCluster{
				svc:  svc,
				ARN:  cluster.ClusterArn,
				Name: cluster.ClusterName,
			})
		}
	}
	return resources, nil
}

type DocDBElasticCluster struct {
	svc *docdbelastic.Client

	ARN  *string
	Name *string
}

func (r *DocDBElasticCluster) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteCluster(ctx, &docdbelastic.DeleteClusterInput{
		ClusterArn: r.ARN,
	})
	return err
}

func (r *DocDBElasticCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DocDBElasticCluster) String() string {
	return *r.Name
}
