package resources

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/dsql"
	dsqltypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DSQLClusterResource = "DSQLCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     DSQLClusterResource,
		Scope:    nuke.Account,
		Resource: &DSQLCluster{},
		Lister:   &DSQLClusterLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type DSQLClusterLister struct{}

func (l *DSQLClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := dsql.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &dsql.ListClustersInput{
		MaxResults: ptr.Int32(100),
	}

	for {
		res, err := svc.ListClusters(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, clusterSummary := range res.Clusters {
			// get additional cluster metadata not returned in ListClusters
			cluster, err := svc.GetCluster(ctx, &dsql.GetClusterInput{
				Identifier: clusterSummary.Identifier,
			})
			if err != nil {
				return nil, err
			}
			// get cluster tags
			var tags map[string]string
			tagsRes, err := svc.ListTagsForResource(ctx, &dsql.ListTagsForResourceInput{
				ResourceArn: clusterSummary.Arn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for dsql cluster: %s", ptr.ToString(clusterSummary.Arn))
			} else {
				tags = tagsRes.Tags
			}

			resources = append(resources, &DSQLCluster{
				svc:                       svc,
				Arn:                       clusterSummary.Arn,
				CreationTime:              cluster.CreationTime,
				DeletionProtectionEnabled: cluster.DeletionProtectionEnabled,
				Identifier:                clusterSummary.Identifier,
				Status:                    cluster.Status,
				Tags:                      tags,
			})
		}

		if res.NextToken == nil {
			break
		}

		params.NextToken = res.NextToken
	}

	return resources, nil
}

func (l *DSQLClusterLister) IsSupportedRegion(region string) bool {
	// https://aws.amazon.com/rds/aurora/dsql/faqs/#:~:text=available%20in%20all-,AWS%20Regions,-%3F
	// NOTE: us-west-2 (Oregon) is available as a witness region, but clusters cannot be created in this region
	supportedRegions := []string{
		"us-east-1",
		"us-east-2",
	}

	return slices.Contains(supportedRegions, region)
}

type DSQLCluster struct {
	svc                       *dsql.Client
	settings                  *libsettings.Setting
	Arn                       *string                 `description:"The ARN of the cluster"`
	CreationTime              *time.Time              `description:"The creation timestamp of the cluster"`
	DeletionProtectionEnabled *bool                   `description:"Boolean indicating cluster deletion prevention"`
	Identifier                *string                 `description:"The identifier of the cluster (eg. iiabt5az32iwdnj4xpxwl5mz3e)"`
	Status                    dsqltypes.ClusterStatus `description:"The status of the cluster at list time"`
	Tags                      map[string]string
}

func (r *DSQLCluster) Remove(ctx context.Context) error {
	err := r.RemoveDeletionProtection(ctx)
	if err != nil {
		return err
	}

	_, err = r.svc.DeleteCluster(ctx, &dsql.DeleteClusterInput{
		Identifier: r.Identifier,
	})

	return err
}

func (r *DSQLCluster) Filter() error {
	if r.Status == dsqltypes.ClusterStatusDeleted {
		return errors.New("dsql cluster already deleted")
	}

	return nil
}

func (r *DSQLCluster) RemoveDeletionProtection(ctx context.Context) error {
	if !r.settings.GetBool("DisableDeletionProtection") {
		return nil
	}

	_, err := r.svc.UpdateCluster(ctx, &dsql.UpdateClusterInput{
		Identifier:                r.Identifier,
		DeletionProtectionEnabled: ptr.Bool(false),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *DSQLCluster) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

func (r *DSQLCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DSQLCluster) String() string {
	return *r.Arn
}
