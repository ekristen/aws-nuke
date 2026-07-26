package resources

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ResolverFirewallRuleGroupAssociationResource = "Route53ResolverFirewallRuleGroupAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ResolverFirewallRuleGroupAssociationResource,
		Scope:    nuke.Account,
		Resource: &Route53ResolverFirewallRuleGroupAssociation{},
		Lister:   &Route53ResolverFirewallRuleGroupAssociationLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type Route53ResolverFirewallRuleGroupAssociationLister struct {
	svc Route53ResolverAPI
}

// List returns all Route53 Resolver Firewall Rule Group to VPC associations in the account.
func (l *Route53ResolverFirewallRuleGroupAssociationLister) List(ctx context.Context, o interface{}) (
	[]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if l.svc == nil {
		l.svc = r53r.NewFromConfig(*opts.Config)
	}

	params := &r53r.ListFirewallRuleGroupAssociationsInput{}
	for {
		resp, err := l.svc.ListFirewallRuleGroupAssociations(ctx, params)
		if err != nil {
			return nil, err
		}

		for i := range resp.FirewallRuleGroupAssociations {
			assoc := resp.FirewallRuleGroupAssociations[i]
			resources = append(resources, &Route53ResolverFirewallRuleGroupAssociation{
				svc:                 l.svc,
				ID:                  assoc.Id,
				Arn:                 assoc.Arn,
				Name:                assoc.Name,
				FirewallRuleGroupID: assoc.FirewallRuleGroupId,
				VpcID:               assoc.VpcId,
				Priority:            assoc.Priority,
				MutationProtection:  assoc.MutationProtection,
				Status:              assoc.Status,
				CreatorRequestID:    assoc.CreatorRequestId,
				ManagedOwnerName:    assoc.ManagedOwnerName,
				CreationTime:        assoc.CreationTime,
				ModificationTime:    assoc.ModificationTime,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// Route53ResolverFirewallRuleGroupAssociation is the resource type for a single FRG to VPC association.
type Route53ResolverFirewallRuleGroupAssociation struct {
	svc                 Route53ResolverAPI
	settings            *libsettings.Setting
	ID                  *string                                      `description:"The ID of the firewall rule group association"`
	Arn                 *string                                      `description:"The ARN of the firewall rule group association"`
	Name                *string                                      `description:"The name of the firewall rule group association"`
	FirewallRuleGroupID *string                                      `description:"The ID of the associated firewall rule group"`
	VpcID               *string                                      `description:"The ID of the associated VPC"`
	Priority            *int32                                       `description:"The processing order of the rule group within the VPC"`
	MutationProtection  r53rtypes.MutationProtectionStatus           `description:"Whether mutation protection is enabled for the association"`
	Status              r53rtypes.FirewallRuleGroupAssociationStatus `description:"The current status of the association"`
	CreatorRequestID    *string                                      `description:"The unique ID for the request that created the association"`
	ManagedOwnerName    *string                                      `description:"The owner of the association, if not managed by you"`
	CreationTime        *string                                      `description:"The time the association was created (Unix time, UTC)"`
	ModificationTime    *string                                      `description:"The time the association was last changed (Unix time, UTC)"`
}

func (r *Route53ResolverFirewallRuleGroupAssociation) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

// Filter excludes associations that are managed by another owner (e.g. Firewall Manager) since
// those cannot be disassociated directly.
func (r *Route53ResolverFirewallRuleGroupAssociation) Filter() error {
	if r.ManagedOwnerName != nil && *r.ManagedOwnerName != "" {
		return errors.New("cannot delete association managed by another owner")
	}
	return nil
}

// Remove implements Resource. It disables mutation protection (when permitted) before
// disassociating the firewall rule group from the VPC. Disassociation is asynchronous.
func (r *Route53ResolverFirewallRuleGroupAssociation) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	if r.settings.GetBool("DisableDeletionProtection") &&
		r.MutationProtection == r53rtypes.MutationProtectionStatusEnabled {
		// Disable mutation protection so the association can be removed.
		// This call is fast and appears to be synchronous.
		_, err := r.svc.UpdateFirewallRuleGroupAssociation(ctx,
			&r53r.UpdateFirewallRuleGroupAssociationInput{
				FirewallRuleGroupAssociationId: r.ID,
				MutationProtection:             r53rtypes.MutationProtectionStatusDisabled,
			},
		)
		if err != nil && !errors.As(err, &notFound) {
			return err
		}
	}

	// Remove the association. This results in an async disassociation which can take some
	// time to complete.
	_, err := r.svc.DisassociateFirewallRuleGroup(ctx, &r53r.DisassociateFirewallRuleGroupInput{
		FirewallRuleGroupAssociationId: r.ID,
	})
	if err != nil {
		// ignore notFound, probably already disassociated
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}

	return nil
}

func (r *Route53ResolverFirewallRuleGroupAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Route53ResolverFirewallRuleGroupAssociation) String() string {
	return *r.ID
}

// HandleWait waits for the disassociation to complete. Disassociation is asynchronous so we poll
// the association status until it is gone (deleted) or no longer in a pending state.
func (r *Route53ResolverFirewallRuleGroupAssociation) HandleWait(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	params := &r53r.ListFirewallRuleGroupAssociationsInput{
		FirewallRuleGroupId: r.FirewallRuleGroupID,
	}

	resp, err := r.svc.ListFirewallRuleGroupAssociations(ctx, params)
	if err != nil {
		// Not found means the association (and its rule group) are gone.
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}

	for i := range resp.FirewallRuleGroupAssociations {
		assoc := resp.FirewallRuleGroupAssociations[i]
		if assoc.Id == nil || *assoc.Id != *r.ID {
			continue
		}

		if assoc.Status == r53rtypes.FirewallRuleGroupAssociationStatusDeleting ||
			assoc.Status == r53rtypes.FirewallRuleGroupAssociationStatusUpdating {
			logrus.Infof("Association %s is in status %s", *r.ID, assoc.Status)
			return liberrors.ErrWaitResource("waiting for association to stabilize")
		}
	}

	// The association is no longer present (or no longer pending), disassociation is complete.
	return nil
}
