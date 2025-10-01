package resources

import (
	"context"

	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EKSClusterResource = "EKSCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     EKSClusterResource,
		Scope:    nuke.Account,
		Resource: &EKSCluster{},
		Lister:   &EKSClusterLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type EKSClusterLister struct{}

func (l *EKSClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := eks.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &eks.ListClustersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := eks.NewListClustersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			dcResp, err := svc.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(cluster)})
			if err != nil {
				return nil, err
			}
			resources = append(resources, &EKSCluster{
				svc:        svc,
				Name:       aws.String(cluster),
				CreatedAt:  dcResp.Cluster.CreatedAt,
				protection: dcResp.Cluster.DeletionProtection,
				Tags:       dcResp.Cluster.Tags,
			})
		}
	}
	return resources, nil
}

type EKSCluster struct {
	Name      *string
	CreatedAt *time.Time
	Tags      map[string]string

	svc        *eks.Client
	settings   *libsettings.Setting
	protection *bool
}

func (r *EKSCluster) Remove(ctx context.Context) error {
	if ptr.ToBool(r.protection) && r.settings.GetBool("DisableDeletionProtection") {
		updateClusterConfigInput := &eks.UpdateClusterConfigInput{
			Name:               r.Name,
			DeletionProtection: aws.Bool(false),
		}
		if _, err := r.svc.UpdateClusterConfig(ctx, updateClusterConfigInput); err != nil {
			return err
		}
	}
	_, err := r.svc.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: r.Name,
	})

	return err
}

func (r *EKSCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EKSCluster) String() string {
	return *r.Name
}

func (r *EKSCluster) Settings(setting *libsettings.Setting) {
	r.settings = setting
}
