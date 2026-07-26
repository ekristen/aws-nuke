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

// Route53ResolverFirewallRuleGroup is the resource type
type Route53ResolverFirewallRuleGroup struct {
	svc              Route53ResolverAPI
	rules            []*Route53ResolverFirewallRule
	Arn              *string               `description:"The ARN of the firewall rule group"`
	CreatorRequestID *string               `description:"The unique identifier (ID) for the request that created the firewall rule group"`
	ID               *string               `description:"The unique identifier (ID) for the firewall rule group"`
	OwnerID          *string               `description:"ID of the AWS account that created the firewall rule group"`
	Name             *string               `description:"Name of the firewall rule group"`
	ShareStatus      r53rtypes.ShareStatus `description:"The current sharing status of the firewall rule group"`
}

// Remove implements Resource
func (r *Route53ResolverFirewallRuleGroup) Remove(ctx context.Context) error {
	var notFound *r53rtypes.ResourceNotFoundException

	// Associations are setup as DependsOn, and should have been removed so all
	// we need to do remove rules which seems to be synchronous or at least very fast
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

	// Delete the FRG
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

// DependsOn ensures that firewall rule group associations are removed before the rule group itself.
func (r *Route53ResolverFirewallRuleGroup) DependsOn() []string {
	return []string{
		Route53ResolverFirewallRuleGroupAssociationResource,
	}
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
