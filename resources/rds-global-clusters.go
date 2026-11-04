package resources

import (
	"context"
	"errors"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/smithy-go"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RDSGlobalClusterResource = "RDSGlobalCluster"

const rdsGlobalClusterStatusDeleting = "deleting"

func init() {
	registry.Register(&registry.Registration{
		Name:     RDSGlobalClusterResource,
		Scope:    nuke.Account,
		Resource: &RDSGlobalCluster{},
		Lister:   &RDSGlobalClusterLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type RDSGlobalClusterLister struct {
	svc RDSAPI
}

// List returns the Aurora global clusters. Their members stay undeletable until the global cluster is gone:
// DeleteDBCluster fails with InvalidDBClusterStateFault for a secondary and InvalidGlobalClusterStateFault for the
// primary.
func (l *RDSGlobalClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := l.svc
	if svc == nil {
		svc = rds.NewFromConfig(*opts.Config)
	}

	paginator := rds.NewDescribeGlobalClustersPaginator(svc, &rds.DescribeGlobalClustersInput{})

	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range res.GlobalClusters {
			globalCluster := res.GlobalClusters[i]

			if !l.belongsToRegion(&globalCluster, opts.Region.Name) {
				continue
			}

			var tags map[string]string
			tagsRes, err := svc.ListTagsForResource(ctx, &rds.ListTagsForResourceInput{
				ResourceName: globalCluster.GlobalClusterArn,
			})
			if err != nil {
				// A global cluster that a run just deleted is not worth a warning; every other error is.
				if !isGlobalClusterNotFound(err) {
					opts.Logger.Warnf("unable to fetch tags for global cluster %s: %v",
						ptr.ToString(globalCluster.GlobalClusterIdentifier), err)
				}
			} else {
				tags = make(map[string]string, len(tagsRes.TagList))
				for _, tag := range tagsRes.TagList {
					tags[ptr.ToString(tag.Key)] = ptr.ToString(tag.Value)
				}
			}

			resources = append(resources, &RDSGlobalCluster{
				svc:                svc,
				Identifier:         globalCluster.GlobalClusterIdentifier,
				Engine:             globalCluster.Engine,
				EngineVersion:      globalCluster.EngineVersion,
				Status:             globalCluster.Status,
				DeletionProtection: globalCluster.DeletionProtection,
				Members:            ptr.Int(len(globalCluster.GlobalClusterMembers)),
				Tags:               tags,
			})
		}
	}

	return resources, nil
}

// belongsToRegion prevents one listing per region: DescribeGlobalClusters returns the same global clusters
// everywhere and their ARN carries no region, so the writer member's region decides. Without a writer every region
// lists it, and the extra removals end in a GlobalClusterNotFoundFault.
func (l *RDSGlobalClusterLister) belongsToRegion(globalCluster *rdstypes.GlobalCluster, region string) bool {
	for _, member := range globalCluster.GlobalClusterMembers {
		if !ptr.ToBool(member.IsWriter) {
			continue
		}

		memberARN, err := arn.Parse(ptr.ToString(member.DBClusterArn))
		if err != nil {
			return true
		}

		return memberARN.Region == region
	}

	return true
}

type RDSGlobalCluster struct {
	svc      RDSAPI
	settings *libsettings.Setting

	// detached holds the members whose removal AWS already accepted; it applies them asynchronously.
	detached map[string]bool

	Identifier         *string           `description:"The identifier of the global cluster"`
	Engine             *string           `description:"The database engine of the global cluster"`
	EngineVersion      *string           `description:"The engine version of the global cluster"`
	Status             *string           `description:"The status of the global cluster at list time"`
	DeletionProtection *bool             `description:"Whether deletion protection is enabled for the global cluster"`
	Members            *int              `description:"The number of clusters attached to the global cluster at list time"`
	Tags               map[string]string `description:"The tags of the global cluster"`
}

// Remove starts the removal. AWS applies member removals asynchronously and rejects the writer while another
// member is attached, so HandleWait drives the remaining steps.
func (r *RDSGlobalCluster) Remove(ctx context.Context) error {
	if err := r.disableDeletionProtection(ctx); err != nil {
		return err
	}

	_, err := r.advance(ctx)

	return err
}

func (r *RDSGlobalCluster) HandleWait(ctx context.Context) error {
	removed, err := r.advance(ctx)
	if err != nil {
		return err
	}

	if !removed {
		return liberrors.ErrWaitResource("waiting for the member clusters to detach")
	}

	return nil
}

func (r *RDSGlobalCluster) Filter() error {
	if ptr.ToString(r.Status) == rdsGlobalClusterStatusDeleting {
		return errors.New("global cluster is already deleting")
	}

	return nil
}

func (r *RDSGlobalCluster) Settings(setting *libsettings.Setting) {
	r.settings = setting
}

func (r *RDSGlobalCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *RDSGlobalCluster) String() string {
	return ptr.ToString(r.Identifier)
}

// advance performs the next removal step and reports whether the global cluster is gone.
func (r *RDSGlobalCluster) advance(ctx context.Context) (bool, error) {
	members, found, err := r.members(ctx)
	if err != nil {
		return false, err
	}

	if !found {
		return true, nil
	}

	pending := readerARNs(members)
	if len(pending) == 0 {
		if writerARN := writerARN(members); writerARN != nil {
			pending = []*string{writerARN}
		}
	}

	if len(pending) == 0 {
		return r.deleteGlobalCluster(ctx)
	}

	for _, memberARN := range pending {
		if err := r.detachMember(ctx, memberARN); err != nil {
			return false, err
		}
	}

	return false, nil
}

// members re-reads the member clusters, because the listed ones are stale as soon as the first member detaches.
func (r *RDSGlobalCluster) members(ctx context.Context) ([]rdstypes.GlobalClusterMember, bool, error) {
	res, err := r.svc.DescribeGlobalClusters(ctx, &rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: r.Identifier,
	})
	if isGlobalClusterNotFound(err) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	if len(res.GlobalClusters) == 0 {
		return nil, false, nil
	}

	return res.GlobalClusters[0].GlobalClusterMembers, true, nil
}

func (r *RDSGlobalCluster) detachMember(ctx context.Context, memberARN *string) error {
	if r.detached[ptr.ToString(memberARN)] {
		return nil
	}

	_, err := r.svc.RemoveFromGlobalCluster(ctx, &rds.RemoveFromGlobalClusterInput{
		GlobalClusterIdentifier: r.Identifier,
		DbClusterIdentifier:     memberARN,
	})
	if err != nil && !isMemberNotFound(err) {
		return err
	}

	if r.detached == nil {
		r.detached = make(map[string]bool)
	}

	r.detached[ptr.ToString(memberARN)] = true

	return nil
}

func (r *RDSGlobalCluster) deleteGlobalCluster(ctx context.Context) (bool, error) {
	_, err := r.svc.DeleteGlobalCluster(ctx, &rds.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: r.Identifier,
	})

	if err == nil || isGlobalClusterNotFound(err) {
		return true, nil
	}

	return false, err
}

func (r *RDSGlobalCluster) disableDeletionProtection(ctx context.Context) error {
	if !ptr.ToBool(r.DeletionProtection) || !r.settings.GetBool("DisableDeletionProtection") {
		return nil
	}

	_, err := r.svc.ModifyGlobalCluster(ctx, &rds.ModifyGlobalClusterInput{
		GlobalClusterIdentifier: r.Identifier,
		DeletionProtection:      ptr.Bool(false),
	})

	return err
}

func readerARNs(members []rdstypes.GlobalClusterMember) []*string {
	var readers []*string

	for _, member := range members {
		if !ptr.ToBool(member.IsWriter) {
			readers = append(readers, member.DBClusterArn)
		}
	}

	return readers
}

func writerARN(members []rdstypes.GlobalClusterMember) *string {
	for _, member := range members {
		if ptr.ToBool(member.IsWriter) {
			return member.DBClusterArn
		}
	}

	return nil
}

func isGlobalClusterNotFound(err error) bool {
	var notFound *rdstypes.GlobalClusterNotFoundFault
	return errors.As(err, &notFound)
}

// isMemberNotFound reports an already detached member. AWS uses a generic InvalidParameterValue for it.
func isMemberNotFound(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.ErrorCode() == "InvalidParameterValue" &&
		strings.Contains(apiErr.ErrorMessage(), "is not found in global cluster")
}
