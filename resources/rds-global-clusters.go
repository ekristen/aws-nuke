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

// rdsGlobalClusterStatusDeleting is the status of a global cluster whose deletion is already running.
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

// List returns the Aurora global clusters of the account. A global cluster has to be removed for its member clusters to
// become deletable: DeleteDBCluster answers with InvalidDBClusterStateFault for a secondary member and with
// InvalidGlobalClusterStateFault for the primary member.
func (l *RDSGlobalClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if l.svc == nil {
		l.svc = rds.NewFromConfig(*opts.Config)
	}

	paginator := rds.NewDescribeGlobalClustersPaginator(l.svc, &rds.DescribeGlobalClustersInput{})

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
			tagsRes, err := l.svc.ListTagsForResource(ctx, &rds.ListTagsForResourceInput{
				ResourceName: globalCluster.GlobalClusterArn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for global cluster %s: %v",
					ptr.ToString(globalCluster.GlobalClusterIdentifier), err)
			} else {
				tags = make(map[string]string, len(tagsRes.TagList))
				for _, tag := range tagsRes.TagList {
					tags[ptr.ToString(tag.Key)] = ptr.ToString(tag.Value)
				}
			}

			resources = append(resources, &RDSGlobalCluster{
				svc:                l.svc,
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

// belongsToRegion reports whether the global cluster is to be listed for the given region. Global clusters are not
// bound to a region, DescribeGlobalClusters returns the same global clusters in every region, and the global cluster
// ARN carries no region either. To list every global cluster exactly once, it is listed for the region of its writer
// member. A global cluster without a writer member has no such region and is listed for every region; the redundant
// removals of the other regions end in a GlobalClusterNotFoundFault, which counts as removed.
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

	// detached holds the member clusters whose removal was already requested, so that an asynchronous removal is not
	// requested again on every wait.
	detached map[string]bool

	Identifier         *string           `description:"The identifier of the global cluster"`
	Engine             *string           `description:"The database engine of the global cluster"`
	EngineVersion      *string           `description:"The engine version of the global cluster"`
	Status             *string           `description:"The status of the global cluster at list time"`
	DeletionProtection *bool             `description:"Whether deletion protection is enabled for the global cluster"`
	Members            *int              `description:"The number of clusters attached to the global cluster at list time"`
	Tags               map[string]string `description:"The tags of the global cluster"`
}

// Remove starts the removal of the global cluster. Member removal is asynchronous, and AWS rejects the removal of the
// writer member while another member is attached, so the removal takes several steps. HandleWait drives the remaining
// ones.
func (r *RDSGlobalCluster) Remove(ctx context.Context) error {
	if err := r.disableDeletionProtection(ctx); err != nil {
		return err
	}

	_, err := r.advance(ctx)

	return err
}

// HandleWait continues the removal until the global cluster is gone.
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

// advance performs the next removal step and reports whether the global cluster is gone. The reader members are
// detached first, the writer member follows once it is the last member, and the empty global cluster is deleted last.
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

// members returns the current member clusters. They are read again on every step, because the ones from the listing
// are outdated as soon as the first member is detached.
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

	switch {
	case err == nil, isGlobalClusterNotFound(err):
		return true, nil
	case isInvalidGlobalClusterState(err):
		// A member removal is still in flight, so the global cluster is not empty yet.
		return false, nil
	default:
		return false, err
	}
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

func isInvalidGlobalClusterState(err error) bool {
	var invalidState *rdstypes.InvalidGlobalClusterStateFault
	return errors.As(err, &invalidState)
}

// isMemberNotFound reports whether the member cluster is already detached from the global cluster. AWS answers this
// case with a generic InvalidParameterValue error instead of a dedicated error type.
func isMemberNotFound(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.ErrorCode() == "InvalidParameterValue" &&
		strings.Contains(apiErr.ErrorMessage(), "is not found in global cluster")
}
