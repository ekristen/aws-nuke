package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EKSNodegroupResource = "EKSNodegroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     EKSNodegroupResource,
		Scope:    nuke.Account,
		Resource: &EKSNodegroup{},
		Lister:   &EKSNodegroupLister{},
		DeprecatedAliases: []string{
			"EKSNodegroups",
		},
	})
}

type EKSNodegroupLister struct{}

func (l *EKSNodegroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := eks.NewFromConfig(*opts.Config)

	var clusterNames []string
	var resources []resource.Resource

	// fetch all cluster names
	clustersPaginator := eks.NewListClustersPaginator(svc, &eks.ListClustersInput{
		MaxResults: aws.Int32(100),
	})
	for clustersPaginator.HasMorePages() {
		resp, err := clustersPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		clusterNames = append(clusterNames, resp.Clusters...)
	}

	// fetch the associated node groups
	for _, clusterName := range clusterNames {
		nodegroupsPaginator := eks.NewListNodegroupsPaginator(svc, &eks.ListNodegroupsInput{
			ClusterName: aws.String(clusterName),
			MaxResults:  aws.Int32(100),
		})
		for nodegroupsPaginator.HasMorePages() {
			resp, err := nodegroupsPaginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, nodegroupName := range resp.Nodegroups {
				descResp, err := svc.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
					ClusterName:   aws.String(clusterName),
					NodegroupName: aws.String(nodegroupName),
				})
				if err != nil {
					return nil, err
				}
				resources = append(resources, &EKSNodegroup{
					svc:       svc,
					nodegroup: descResp.Nodegroup,
				})
			}
		}
	}

	return resources, nil
}

type EKSNodegroup struct {
	svc       *eks.Client
	nodegroup *ekstypes.Nodegroup
}

func (ng *EKSNodegroup) Remove(ctx context.Context) error {
	_, err := ng.svc.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   ng.nodegroup.ClusterName,
		NodegroupName: ng.nodegroup.NodegroupName,
	})
	return err
}

func (ng *EKSNodegroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Cluster", ng.nodegroup.ClusterName)
	properties.Set("Profile", ng.nodegroup.NodegroupName)
	if ng.nodegroup.CreatedAt != nil {
		properties.Set("CreatedAt", ng.nodegroup.CreatedAt.Format(time.RFC3339))
	}
	for k, v := range ng.nodegroup.Tags {
		properties.SetTag(&k, v)
	}
	return properties
}

func (ng *EKSNodegroup) String() string {
	return fmt.Sprintf("%s:%s", *ng.nodegroup.ClusterName, *ng.nodegroup.NodegroupName)
}
