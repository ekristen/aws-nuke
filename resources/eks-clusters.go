package resources

import (
	"context"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EKSClusterResource = "EKSCluster"

func init() {
	resource.Register(resource.Registration{
		Name:   EKSClusterResource,
		Scope:  nuke.Account,
		Lister: &EKSClusterLister{},
	})
}

type EKSClusterLister struct{}

func (l *EKSClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := eks.New(opts.Session)
	var resources []resource.Resource

	params := &eks.ListClustersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			dcResp, err := svc.DescribeCluster(&eks.DescribeClusterInput{Name: cluster})
			if err != nil {
				return nil, err
			}
			resources = append(resources, &EKSCluster{
				svc:     svc,
				name:    cluster,
				cluster: dcResp.Cluster,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type EKSCluster struct {
	svc     *eks.EKS
	name    *string
	cluster *eks.Cluster
}

func (f *EKSCluster) Remove(_ context.Context) error {

	_, err := f.svc.DeleteCluster(&eks.DeleteClusterInput{
		Name: f.name,
	})

	return err
}

func (f *EKSCluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("CreatedAt", f.cluster.CreatedAt.Format(time.RFC3339))
	for key, value := range f.cluster.Tags {
		properties.SetTag(&key, value)
	}
	return properties
}

func (f *EKSCluster) String() string {
	return *f.name
}
