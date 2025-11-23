package resources

import (
	"context"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
)

// From https://stackoverflow.com/questions/72235425/simplifying-aws-sdk-go-v2-testing-mocking
type Route53ResolverAPI interface {
	DeleteFirewallRule(ctx context.Context, params *r53r.DeleteFirewallRuleInput,
		optFns ...func(*r53r.Options)) (*r53r.DeleteFirewallRuleOutput, error)
	DeleteFirewallDomainList(ctx context.Context, params *r53r.DeleteFirewallDomainListInput,
		optFns ...func(*r53r.Options)) (*r53r.DeleteFirewallDomainListOutput, error)
	DeleteFirewallRuleGroup(ctx context.Context, params *r53r.DeleteFirewallRuleGroupInput,
		optFns ...func(*r53r.Options)) (*r53r.DeleteFirewallRuleGroupOutput, error)
	DisassociateFirewallRuleGroup(ctx context.Context, params *r53r.DisassociateFirewallRuleGroupInput,
		optFns ...func(*r53r.Options)) (*r53r.DisassociateFirewallRuleGroupOutput, error)
	ListFirewallDomainLists(ctx context.Context, params *r53r.ListFirewallDomainListsInput,
		optFns ...func(*r53r.Options)) (*r53r.ListFirewallDomainListsOutput, error)
	ListFirewallRuleGroups(ctx context.Context, params *r53r.ListFirewallRuleGroupsInput,
		optFns ...func(*r53r.Options)) (*r53r.ListFirewallRuleGroupsOutput, error)
	ListFirewallRuleGroupAssociations(ctx context.Context, params *r53r.ListFirewallRuleGroupAssociationsInput,
		optFns ...func(*r53r.Options)) (*r53r.ListFirewallRuleGroupAssociationsOutput, error)
	ListFirewallRules(ctx context.Context, params *r53r.ListFirewallRulesInput,
		optFns ...func(*r53r.Options)) (*r53r.ListFirewallRulesOutput, error)
	DeleteResolverQueryLogConfig(ctx context.Context, params *r53r.DeleteResolverQueryLogConfigInput,
		optFns ...func(*r53r.Options)) (*r53r.DeleteResolverQueryLogConfigOutput, error)
	DisassociateResolverQueryLogConfig(ctx context.Context, params *r53r.DisassociateResolverQueryLogConfigInput,
		optFns ...func(*r53r.Options)) (*r53r.DisassociateResolverQueryLogConfigOutput, error)
	ListResolverQueryLogConfigs(ctx context.Context, params *r53r.ListResolverQueryLogConfigsInput,
		optFns ...func(*r53r.Options)) (*r53r.ListResolverQueryLogConfigsOutput, error)
	ListResolverQueryLogConfigAssociations(ctx context.Context,
		params *r53r.ListResolverQueryLogConfigAssociationsInput,
		optFns ...func(*r53r.Options)) (*r53r.ListResolverQueryLogConfigAssociationsOutput, error)
}

type Route53ResolverClient struct {
	Client *r53r.Client
}

func (c *Route53ResolverClient) DeleteFirewallRule(ctx context.Context, params *r53r.DeleteFirewallRuleInput,
	optFns ...func(*r53r.Options)) (*r53r.DeleteFirewallRuleOutput, error) {
	return c.Client.DeleteFirewallRule(ctx, params, optFns...)
}

func (c *Route53ResolverClient) DeleteFirewallDomainList(ctx context.Context,
	params *r53r.DeleteFirewallDomainListInput,
	optFns ...func(*r53r.Options)) (*r53r.DeleteFirewallDomainListOutput, error) {
	return c.Client.DeleteFirewallDomainList(ctx, params, optFns...)
}

func (c *Route53ResolverClient) DeleteFirewallRuleGroup(ctx context.Context,
	params *r53r.DeleteFirewallRuleGroupInput,
	optFns ...func(*r53r.Options)) (*r53r.DeleteFirewallRuleGroupOutput, error) {
	return c.Client.DeleteFirewallRuleGroup(ctx, params, optFns...)
}

func (c *Route53ResolverClient) DisassociateFirewallRuleGroup(ctx context.Context,
	params *r53r.DisassociateFirewallRuleGroupInput,
	optFns ...func(*r53r.Options)) (*r53r.DisassociateFirewallRuleGroupOutput, error) {
	return c.Client.DisassociateFirewallRuleGroup(ctx, params, optFns...)
}

func (c *Route53ResolverClient) ListFirewallRuleGroupAssociations(ctx context.Context,
	params *r53r.ListFirewallRuleGroupAssociationsInput,
	optFns ...func(*r53r.Options)) (*r53r.ListFirewallRuleGroupAssociationsOutput, error) {
	return c.Client.ListFirewallRuleGroupAssociations(ctx, params, optFns...)
}

func (c *Route53ResolverClient) ListFirewallDomainLists(ctx context.Context,
	params *r53r.ListFirewallDomainListsInput,
	optFns ...func(*r53r.Options)) (*r53r.ListFirewallDomainListsOutput, error) {
	return c.Client.ListFirewallDomainLists(ctx, params, optFns...)
}

func (c *Route53ResolverClient) ListFirewallRuleGroups(ctx context.Context,
	params *r53r.ListFirewallRuleGroupsInput,
	optFns ...func(*r53r.Options)) (*r53r.ListFirewallRuleGroupsOutput, error) {
	return c.Client.ListFirewallRuleGroups(ctx, params, optFns...)
}

func (c *Route53ResolverClient) DeleteResolverQueryLogConfig(ctx context.Context,
	params *r53r.DeleteResolverQueryLogConfigInput,
	optFns ...func(*r53r.Options)) (*r53r.DeleteResolverQueryLogConfigOutput, error) {
	return c.Client.DeleteResolverQueryLogConfig(ctx, params, optFns...)
}

func (c *Route53ResolverClient) DisassociateResolverQueryLogConfig(ctx context.Context,
	params *r53r.DisassociateResolverQueryLogConfigInput,
	optFns ...func(*r53r.Options)) (*r53r.DisassociateResolverQueryLogConfigOutput, error) {
	return c.Client.DisassociateResolverQueryLogConfig(ctx, params, optFns...)
}

func (c *Route53ResolverClient) ListFirewallRules(ctx context.Context,
	params *r53r.ListFirewallRulesInput,
	optFns ...func(*r53r.Options)) (*r53r.ListFirewallRulesOutput, error) {
	return c.Client.ListFirewallRules(ctx, params, optFns...)
}

func (c *Route53ResolverClient) ListResolverQueryLogConfigAssociations(ctx context.Context,
	params *r53r.ListResolverQueryLogConfigAssociationsInput,
	optFns ...func(*r53r.Options)) (*r53r.ListResolverQueryLogConfigAssociationsOutput, error) {
	return c.Client.ListResolverQueryLogConfigAssociations(ctx, params, optFns...)
}

func (c *Route53ResolverClient) ListResolverQueryLogConfigs(ctx context.Context,
	params *r53r.ListResolverQueryLogConfigsInput,
	optFns ...func(*r53r.Options)) (*r53r.ListResolverQueryLogConfigsOutput, error) {
	return c.Client.ListResolverQueryLogConfigs(ctx, params, optFns...)
}
