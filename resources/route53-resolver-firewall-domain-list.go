package resources

import (
	"context"
	"fmt"

	r53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53ResolverFirewallDomainListResource = "Route53ResolverFirewallDomainList"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53ResolverFirewallDomainListResource,
		Scope:    nuke.Account,
		Resource: &Route53ResolverFirewallDomainList{},
		Lister:   &Route53ResolverFirewallDomainListLister{},
	})
}

type Route53ResolverFirewallDomainListLister struct {
	svc Route53ResolverAPI
}

// List returns a list of all Route53 Resolver Firewall DomainLists before filtering to be nuked
func (l *Route53ResolverFirewallDomainListLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource

	if l.svc == nil {
		opts := o.(*nuke.ListerOpts)
		svc := r53r.NewFromConfig(*opts.Config)
		l.svc = svc
	}

	params := &r53r.ListFirewallDomainListsInput{}
	for {
		resp, err := l.svc.ListFirewallDomainLists(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, domainList := range resp.FirewallDomainLists {
			resources = append(resources, &Route53ResolverFirewallDomainList{
				svc:              l.svc,
				Arn:              domainList.Arn,
				CreatorRequestId: domainList.CreatorRequestId,
				Id:               domainList.Id,
				ManagedOwnerName: domainList.ManagedOwnerName,
				Name:             domainList.Name,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

// Route53ResolverFirewallDomainList is the resource type
type Route53ResolverFirewallDomainList struct {
	svc              Route53ResolverAPI
	Arn              *string
	CreatorRequestId *string
	Id               *string
	ManagedOwnerName *string
	Name             *string
}

func (r *Route53ResolverFirewallDomainList) Filter() error {
	// Domain lists created by AWS will have a ManagedOwnerName set
	if r.ManagedOwnerName != nil && *r.ManagedOwnerName != "" {
		return fmt.Errorf("cannot delete system defined domain lists")
	}

	return nil
}

func (r *Route53ResolverFirewallDomainList) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteFirewallDomainList(ctx, &r53r.DeleteFirewallDomainListInput{
		FirewallDomainListId: r.Id,
	})

	return err
}

func (r *Route53ResolverFirewallDomainList) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
