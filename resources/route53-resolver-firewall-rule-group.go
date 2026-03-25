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

const Route53ResolverFirewallRuleGroupResource = "Route53ResolverFirewallRuleGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ResolverFirewallRuleGroupResource,
		Scope:    nuke.Account,
		Resource: &Route53ResolverFirewallRuleGroup{},
		Lister:   &Route53ResolverFirewallRuleGroupLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type Route53ResolverFirewallRuleGroupLister struct {
	svc Route53ResolverAPI
}

// List returns a list of all Route53 Resolver Firewall RuleGroups before filtering to be nuked
func (l *Route53ResolverFirewallRuleGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if l.svc == nil {
		l.svc = r53r.NewFromConfig(*opts.Config)
	}

	vpcAssociations, vpcErr := ruleGroupsToAssociationIds(ctx, l.svc)
	if vpcErr != nil {
		return nil, vpcErr
	}

	params := &r53r.ListFirewallRuleGroupsInput{}
	for {
		resp, err := l.svc.ListFirewallRuleGroups(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, firewallRuleGroup := range resp.FirewallRuleGroups {
			firewallRules, ruleErr := getFirewallRules(ctx, l.svc, firewallRuleGroup.Id)
			if ruleErr != nil {
				return nil, ruleErr
			}

			resources = append(resources, &Route53ResolverFirewallRuleGroup{
				svc:              l.svc,
				vpcAssociations:  vpcAssociations[*firewallRuleGroup.Id],
				rules:            firewallRules,
				Arn:              firewallRuleGroup.Arn,
				CreatorRequestID: firewallRuleGroup.CreatorRequestId,
				ID:               firewallRuleGroup.Id,
				OwnerID:          firewallRuleGroup.OwnerId,
				Name:             firewallRuleGroup.Name,
				ShareStatus:      firewallRuleGroup.ShareStatus,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// Fields in Firewall Rule we need to know for deletes
type Route53ResolverFirewallRule struct {
	Name                       *string
	FirewallDomainListID       *string
	Qtype                      *string
	FirewallThreatProtectionID *string
}

type Route53ResolverFirewallRuleGroupVpcAssociation struct {
	ID                 *string
	MutationProtection r53rtypes.MutationProtectionStatus
}

// Route53ResolverFirewallRuleGroup is the resource type
type Route53ResolverFirewallRuleGroup struct {
	svc              Route53ResolverAPI
	settings         *libsettings.Setting
	vpcAssociations  []*Route53ResolverFirewallRuleGroupVpcAssociation
	rules            []*Route53ResolverFirewallRule
	Arn              *string               `description:"The ARN of the firewall rule group"`
	CreatorRequestID *string               `description:" The unique identifier (ID) for the request that created the firewall rule group"`
	ID               *string               `description:" The unique identifier (ID) for the firewall rule group"`
	OwnerID          *string               `description:" ID of the AWS account that created the firewall rule group"`
	Name             *string               `description:" Name of the firewall rule group"`
	ShareStatus      r53rtypes.ShareStatus `description:" The current sharing status of the firewall rule group"`
}

func (r *Route53ResolverFirewallRuleGroup) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

// Remove implements Resource
func (r *Route53ResolverFirewallRuleGroup) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	// disassociate VPCs first since that's slower
	for _, vpcAssociation := range r.vpcAssociations {
		if r.settings.GetBool("DisableDeletionProtection") && vpcAssociation.MutationProtection == r53rtypes.MutationProtectionStatusEnabled {
			// disable mutation protection for any associations that have it enabled
			// This call is very fast and seems to be synchronous
			_, err := r.svc.UpdateFirewallRuleGroupAssociation(ctx,
				&r53r.UpdateFirewallRuleGroupAssociationInput{
					FirewallRuleGroupAssociationId: vpcAssociation.ID,
					MutationProtection:             r53rtypes.MutationProtectionStatusDisabled,
				},
			)
			// ignore, probably already disassociated
			if errors.As(err, &notFound) {
				continue
			}
		}

		// Remove the association.  This call results in an async dissociation which can
		// take some time to complete
		_, err := r.svc.DisassociateFirewallRuleGroup(ctx, &r53r.DisassociateFirewallRuleGroupInput{
			FirewallRuleGroupAssociationId: vpcAssociation.ID,
		})
		if err != nil {
			// ignore notFound, probably already disassociated
			if errors.As(err, &notFound) {
				continue
			}
			return err
		}
	}

	// then remove rules
	for _, rule := range r.rules {
		_, err := r.svc.DeleteFirewallRule(ctx, &r53r.DeleteFirewallRuleInput{
			FirewallRuleGroupId:        r.ID,
			FirewallDomainListId:       rule.FirewallDomainListID,
			FirewallThreatProtectionId: rule.FirewallThreatProtectionID,
			Qtype:                      rule.Qtype,
		})

		if err != nil {
			// ignore, rule has probably been deleted
			if errors.As(err, &notFound) {
				continue
			}
			return err
		}
	}

	err := waitForAssociationToStabilize(ctx, r.svc, r)
	if err != nil {
		return err
	}

	// finally delete the FRG
	_, err = r.svc.DeleteFirewallRuleGroup(ctx, &r53r.DeleteFirewallRuleGroupInput{
		FirewallRuleGroupId: r.ID,
	})

	return err
}

func (r *Route53ResolverFirewallRuleGroup) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	// TODO(v4): remove backward-compat properties
	props.Set("Id", r.ID)
	props.Set("CreatorRequestId", r.CreatorRequestID)
	props.Set("OwnerId", r.OwnerID)
	return props
}

func (r *Route53ResolverFirewallRuleGroup) String() string {
	return *r.ID
}

// ruleGroupsToAssociationIds - Associate all the FRG association ids to their firewall rule group ID to be
// disassociated before deleting the rule.
func ruleGroupsToAssociationIds(ctx context.Context, svc Route53ResolverAPI) (map[string][]*Route53ResolverFirewallRuleGroupVpcAssociation,
	error) {
	vpcAssociations := map[string][]*Route53ResolverFirewallRuleGroupVpcAssociation{}

	params := &r53r.ListFirewallRuleGroupAssociationsInput{}

	for {
		// Lists ALL FRG->VPC associations for all FRGs
		resp, err := svc.ListFirewallRuleGroupAssociations(ctx, params)

		if err != nil {
			return nil, err
		}

		frgas := resp.FirewallRuleGroupAssociations
		for i := range frgas {
			associationID := frgas[i].Id
			if associationID != nil {
				frgID := *frgas[i].FirewallRuleGroupId

				frgAssoc := Route53ResolverFirewallRuleGroupVpcAssociation{
					ID:                 associationID,
					MutationProtection: frgas[i].MutationProtection,
				}

				if _, ok := vpcAssociations[frgID]; !ok {
					associations := []*Route53ResolverFirewallRuleGroupVpcAssociation{&frgAssoc}
					vpcAssociations[frgID] = associations
				} else {
					vpcAssociations[frgID] = append(vpcAssociations[frgID], &frgAssoc)
				}
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return vpcAssociations, nil
}

// Get Firewall rules for the FRG with given firewallRuleGroupID
func getFirewallRules(ctx context.Context, svc Route53ResolverAPI, firewallRuleGroupID *string) ([]*Route53ResolverFirewallRule, error) {
	rules := []*Route53ResolverFirewallRule{}

	params := &r53r.ListFirewallRulesInput{
		FirewallRuleGroupId: firewallRuleGroupID,
	}

	for {
		resp, err := svc.ListFirewallRules(ctx, params)

		if err != nil {
			return nil, err
		}

		for i := range resp.FirewallRules {
			rule := Route53ResolverFirewallRule{
				Name:                       resp.FirewallRules[i].Name,
				FirewallDomainListID:       resp.FirewallRules[i].FirewallDomainListId,
				FirewallThreatProtectionID: resp.FirewallRules[i].FirewallThreatProtectionId,
				Qtype:                      resp.FirewallRules[i].Qtype,
			}

			rules = append(rules, &rule)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return rules, nil
}

// Wait for the FRG-to-VPC association to stabilize.
func waitForAssociationToStabilize(ctx context.Context, svc Route53ResolverAPI, frg *Route53ResolverFirewallRuleGroup) error {
	params := &r53r.ListFirewallRuleGroupAssociationsInput{
		FirewallRuleGroupId: frg.ID,
	}

	resp, err := svc.ListFirewallRuleGroupAssociations(ctx, params)
	frgas := resp.FirewallRuleGroupAssociations

	// Not found means we successfully deleted this FRG
	var notFound *r53rtypes.ResourceNotFoundException
	if errors.As(err, &notFound) {
		return err
	}

	statusIsPending := false
	for i := range frgas {
		currentStatus := frgas[i].Status
		associationID := frgas[i].Id

		if currentStatus == r53rtypes.FirewallRuleGroupAssociationStatusUpdating ||
			currentStatus == r53rtypes.FirewallRuleGroupAssociationStatusDeleting {
			logrus.Infof("Association %s on firewall rule group %s is in status %s",
				*associationID, *frg.ID, currentStatus)
			statusIsPending = true
			break
		}
	}

	if statusIsPending {
		// Return an ErrHoldResource to put the resource in ItemStateHold and retry later
		return liberrors.ErrHoldResource("waiting for associations to stabilize")
	}

	return nil
}
