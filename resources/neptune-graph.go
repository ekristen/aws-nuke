package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	neptunegraphtypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NeptuneGraphResource = "NeptuneGraph"

func init() {
	registry.Register(&registry.Registration{
		Name:     NeptuneGraphResource,
		Scope:    nuke.Account,
		Resource: &NeptuneGraph{},
		Lister:   &NeptuneGraphLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type NeptuneGraphLister struct{}

func (l *NeptuneGraphLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := neptunegraph.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	var nextToken *string
	for {
		res, err := svc.ListGraphs(ctx, &neptunegraph.ListGraphsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, p := range res.Graphs {
			// get tags for graph
			t, err := svc.ListTagsForResource(ctx, &neptunegraph.ListTagsForResourceInput{
				ResourceArn: p.Arn,
			})
			if err != nil {
				return nil, err
			}

			snapshots := make([]neptunegraphtypes.GraphSnapshotSummary, 0)
			var snapshotNextToken *string
			for {
				s, err := svc.ListGraphSnapshots(ctx, &neptunegraph.ListGraphSnapshotsInput{
					GraphIdentifier: p.Id,
					NextToken:       snapshotNextToken,
				})
				if err != nil {
					return nil, err
				}
				snapshots = append(snapshots, s.GraphSnapshots...)
				if s.NextToken == nil {
					break
				}
				snapshotNextToken = s.NextToken
			}

			resources = append(resources, &NeptuneGraph{
				svc:       svc,
				ID:        p.Id,
				Arn:       p.Arn,
				Name:      p.Name,
				Status:    (*string)(&p.Status),
				Tags:      t.Tags,
				snapshots: snapshots,
			})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources, nil
}

type NeptuneGraph struct {
	svc       *neptunegraph.Client
	settings  *libsettings.Setting
	ID        *string `description:"The Neptune Graph identifier (e.g. g-prz5mldixa)"`
	Arn       *string `description:"The Neptune Graph resource ARN"`
	Name      *string `description:"The name of the Neptune Graph"`
	Status    *string `description:"The status of the Neptune Graph (e.g. Available/Deleting/Updating)"`
	Tags      map[string]string
	snapshots []neptunegraphtypes.GraphSnapshotSummary
}

func (r *NeptuneGraph) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

func (r *NeptuneGraph) Filter() error {
	if ptr.ToString(r.Status) == "DELETING" {
		return fmt.Errorf("already deleting")
	}
	return nil
}

func (r *NeptuneGraph) Remove(ctx context.Context) error {
	if r.settings.GetBool("DisableDeletionProtection") {
		_, err := r.svc.UpdateGraph(ctx, &neptunegraph.UpdateGraphInput{
			GraphIdentifier:    r.ID,
			DeletionProtection: ptr.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	for _, s := range r.snapshots {
		_, err := r.svc.DeleteGraphSnapshot(ctx, &neptunegraph.DeleteGraphSnapshotInput{
			SnapshotIdentifier: s.Id,
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteGraph(ctx, &neptunegraph.DeleteGraphInput{
		GraphIdentifier: r.ID,
		SkipSnapshot:    ptr.Bool(true),
	})
	return err
}

func (r *NeptuneGraph) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
