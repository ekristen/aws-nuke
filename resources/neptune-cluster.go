package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NeptuneClusterResource = "NeptuneCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     NeptuneClusterResource,
		Scope:    nuke.Account,
		Resource: &NeptuneCluster{},
		Lister:   &NeptuneClusterLister{},
		DependsOn: []string{
			NeptuneInstanceResource,
		},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type NeptuneClusterLister struct{}

func (l *NeptuneClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := neptune.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &neptune.DescribeDBClustersInput{
		Filters: []*neptune.Filter{
			{
				Name:   aws.String("engine"),
				Values: []*string{aws.String("neptune")},
			},
		},
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDBClusters(params)
		if err != nil {
			return nil, err
		}

		for _, dbCluster := range output.DBClusters {
			var dbTags []*neptune.Tag
			tags, err := svc.ListTagsForResource(&neptune.ListTagsForResourceInput{
				ResourceName: dbCluster.DBClusterArn,
			})
			if err != nil {
				opts.Logger.WithError(err).Warn("failed to list tags for resource")
			} else {
				dbTags = tags.TagList
			}

			resources = append(resources, &NeptuneCluster{
				svc:    svc,
				ID:     dbCluster.DBClusterIdentifier,
				Status: dbCluster.Status,
				Tags:   dbTags,
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
	svc      *neptune.Neptune
	settings *libsettings.Setting

	ID     *string
	Status *string
	Tags   []*neptune.Tag
}

func (r *NeptuneCluster) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

func (r *NeptuneCluster) Filter() error {
	if ptr.ToString(r.Status) == "deleting" {
		return fmt.Errorf("already deleting")
	}
	return nil
}

func (r *NeptuneCluster) Remove(_ context.Context) error {
	if r.settings.GetBool("DisableDeletionProtection") {
		_, err := r.svc.ModifyDBCluster(&neptune.ModifyDBClusterInput{
			DBClusterIdentifier: r.ID,
			DeletionProtection:  ptr.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteDBCluster(&neptune.DeleteDBClusterInput{
		DBClusterIdentifier: r.ID,
		SkipFinalSnapshot:   ptr.Bool(true),
	})

	return err
}

func (r *NeptuneCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NeptuneCluster) String() string {
	return *r.ID
}
