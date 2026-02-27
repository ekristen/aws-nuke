package resources

import (
	"context"
	"errors"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	r53rtypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
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
				svc:               l.svc,
				vpcAssociationIds: vpcAssociations[*firewallRuleGroup.Id],
				rules:             firewallRules,
				Arn:               firewallRuleGroup.Arn,
				CreatorRequestID:  firewallRuleGroup.CreatorRequestId,
				ID:                firewallRuleGroup.Id,
				OwnerID:           firewallRuleGroup.OwnerId,
				Name:              firewallRuleGroup.Name,
				ShareStatus:       firewallRuleGroup.ShareStatus,
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

// Route53ResolverFirewallRuleGroup is the resource type
type Route53ResolverFirewallRuleGroup struct {
	svc               Route53ResolverAPI
	vpcAssociationIds []*string
	rules             []*Route53ResolverFirewallRule
	Arn               *string
	CreatorRequestID  *string
	ID                *string
	OwnerID           *string
	Name              *string
	ShareStatus       r53rtypes.ShareStatus
}

// Remove implements Resource
func (r *Route53ResolverFirewallRuleGroup) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	// disassociate VPCs first since that's slower
	for _, associationID := range r.vpcAssociationIds {
		_, err := r.svc.DisassociateFirewallRuleGroup(ctx, &r53r.DisassociateFirewallRuleGroupInput{
			FirewallRuleGroupAssociationId: associationID,
		})
		if err != nil {
			// ignore, probably already associated
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

	// finally delete the FRG
	_, err := r.svc.DeleteFirewallRuleGroup(ctx, &r53r.DeleteFirewallRuleGroupInput{
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
func ruleGroupsToAssociationIds(ctx context.Context, svc Route53ResolverAPI) (map[string][]*string, error) {
	vpcAssociations := map[string][]*string{}

	params := &r53r.ListFirewallRuleGroupAssociationsInput{}

	for {
		resp, err := svc.ListFirewallRuleGroupAssociations(ctx, params)

		if err != nil {
			return nil, err
		}

		frgas := resp.FirewallRuleGroupAssociations
		for i := range frgas {
			associationID := frgas[i].Id
			if associationID != nil {
				frgID := *frgas[i].FirewallRuleGroupId

				if _, ok := vpcAssociations[frgID]; !ok {
					vpcAssociations[frgID] = []*string{associationID}
				} else {
					vpcAssociations[frgID] = append(vpcAssociations[frgID], associationID)
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
