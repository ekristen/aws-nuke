package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/route53resolver"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ResolverRuleResource = "Route53ResolverRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   Route53ResolverRuleResource,
		Scope:  nuke.Account,
		Lister: &Route53ResolverRuleLister{},
	})
}

type Route53ResolverRuleLister struct{}

// List returns a list of all Route53 ResolverRules before filtering to be nuked
func (l *Route53ResolverRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := route53resolver.New(opts.Session)

	vpcAssociations, vpcErr := resolverRulesToVpcIDs(svc)
	if vpcErr != nil {
		return nil, vpcErr
	}

	params := &route53resolver.ListResolverRulesInput{}
	for {
		resp, err := svc.ListResolverRules(params)

		if err != nil {
			return nil, err
		}

		for _, rule := range resp.ResolverRules {
			resources = append(resources, &Route53ResolverRule{
				svc:        svc,
				vpcIds:     vpcAssociations[*rule.Id],
				ID:         rule.Id,
				Name:       rule.Name,
				DomainName: rule.DomainName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// resolverRulesToVpcIDs - Associate all the vpcIDs to their resolver rule ID to be disassociated before deleting the rule.
func resolverRulesToVpcIDs(svc *route53resolver.Route53Resolver) (map[string][]*string, error) {
	vpcAssociations := map[string][]*string{}

	params := &route53resolver.ListResolverRuleAssociationsInput{}

	for {
		resp, err := svc.ListResolverRuleAssociations(params)

		if err != nil {
			return nil, err
		}

		for _, ruleAssociation := range resp.ResolverRuleAssociations {
			vpcID := ruleAssociation.VPCId
			if vpcID != nil {
				resolverRuleID := *ruleAssociation.ResolverRuleId

				if _, ok := vpcAssociations[resolverRuleID]; !ok {
					vpcAssociations[resolverRuleID] = []*string{vpcID}
				} else {
					vpcAssociations[resolverRuleID] = append(vpcAssociations[resolverRuleID], vpcID)
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

// Route53ResolverRule is the resource type
type Route53ResolverRule struct {
	svc        *route53resolver.Route53Resolver
	vpcIds     []*string
	ID         *string
	Name       *string
	DomainName *string
}

// Filter removes resources automatically from being nuked
func (r *Route53ResolverRule) Filter() error {
	if strings.HasPrefix(ptr.ToString(r.ID), "rslvr-autodefined-rr") {
		return fmt.Errorf("cannot delete system defined rules")
	}

	// TODO: is this needed if the system defined is already filtered?
	if r.DomainName != nil && ptr.ToString(r.DomainName) == "." {
		return fmt.Errorf(`filtering DomainName "."`)
	}

	return nil
}

// Remove implements Resource
func (r *Route53ResolverRule) Remove(_ context.Context) error {
	for _, vpcID := range r.vpcIds {
		_, err := r.svc.DisassociateResolverRule(&route53resolver.DisassociateResolverRuleInput{
			ResolverRuleId: r.ID,
			VPCId:          vpcID,
		})

		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteResolverRule(&route53resolver.DeleteResolverRuleInput{
		ResolverRuleId: r.ID,
	})

	return err
}

// Properties provides debugging output
func (r *Route53ResolverRule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

// String implements Stringer
func (r *Route53ResolverRule) String() string {
	return fmt.Sprintf("%s (%s)", ptr.ToString(r.ID), ptr.ToString(r.Name))
}
