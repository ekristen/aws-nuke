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
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc := r53r.NewFromConfig(*opts.Config)
		l.svc = svc
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
				CreatorRequestId:  firewallRuleGroup.CreatorRequestId,
				Id:                firewallRuleGroup.Id,
				OwnerId:           firewallRuleGroup.OwnerId,
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
	FirewallDomainListId       *string
	Qtype                      *string
	FirewallThreatProtectionId *string
}

// Route53ResolverFirewallRuleGroup is the resource type
type Route53ResolverFirewallRuleGroup struct {
	svc               Route53ResolverAPI
	vpcAssociationIds []*string
	rules             []*Route53ResolverFirewallRule
	Arn               *string
	CreatorRequestId  *string
	Id                *string
	OwnerId           *string
	Name              *string
	ShareStatus       r53rtypes.ShareStatus
}

// Remove implements Resource
func (r *Route53ResolverFirewallRuleGroup) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	// disassociate VPCs first since that's slower
	for _, associationId := range r.vpcAssociationIds {
		_, err := r.svc.DisassociateFirewallRuleGroup(ctx, &r53r.DisassociateFirewallRuleGroupInput{
			FirewallRuleGroupAssociationId: associationId,
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
			FirewallRuleGroupId:        r.Id,
			FirewallDomainListId:       rule.FirewallDomainListId,
			FirewallThreatProtectionId: rule.FirewallThreatProtectionId,
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
		FirewallRuleGroupId: r.Id,
	})

	return err
}

func (r *Route53ResolverFirewallRuleGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
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
			associationId := frgas[i].Id
			if associationId != nil {
				frgId := *frgas[i].FirewallRuleGroupId

				if _, ok := vpcAssociations[frgId]; !ok {
					vpcAssociations[frgId] = []*string{associationId}
				} else {
					vpcAssociations[frgId] = append(vpcAssociations[frgId], associationId)
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

// Get Firewall rules for the FRG with given firewallRuleGroupId
func getFirewallRules(ctx context.Context, svc Route53ResolverAPI, firewallRuleGroupId *string) ([]*Route53ResolverFirewallRule, error) {
	rules := []*Route53ResolverFirewallRule{}

	params := &r53r.ListFirewallRulesInput{
		FirewallRuleGroupId: firewallRuleGroupId,
	}

	for {
		resp, err := svc.ListFirewallRules(ctx, params)

		if err != nil {
			return nil, err
		}

		for i := range resp.FirewallRules {
			rule := Route53ResolverFirewallRule{
				Name:                       resp.FirewallRules[i].Name,
				FirewallDomainListId:       resp.FirewallRules[i].FirewallDomainListId,
				FirewallThreatProtectionId: resp.FirewallRules[i].FirewallThreatProtectionId,
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
