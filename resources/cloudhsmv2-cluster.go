package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudHSMV2ClusterResource = "CloudHSMV2Cluster"

func init() {
	resource.Register(resource.Registration{
		Name:   CloudHSMV2ClusterResource,
		Scope:  nuke.Account,
		Lister: &CloudHSMV2ClusterLister{},
	})
}

type CloudHSMV2ClusterLister struct{}

func (l *CloudHSMV2ClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudhsmv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudhsmv2.DescribeClustersInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.DescribeClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			resources = append(resources, &CloudHSMV2Cluster{
				svc:       svc,
				clusterID: cluster.ClusterId,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudHSMV2Cluster struct {
	svc       *cloudhsmv2.CloudHSMV2
	clusterID *string
}

func (f *CloudHSMV2Cluster) Remove(_ context.Context) error {
	_, err := f.svc.DeleteCluster(&cloudhsmv2.DeleteClusterInput{
		ClusterId: f.clusterID,
	})

	return err
}

func (f *CloudHSMV2Cluster) String() string {
	return *f.clusterID
}
