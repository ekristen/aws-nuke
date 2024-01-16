package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const RDSProxyResource = "RDSProxy"

func init() {
	resource.Register(resource.Registration{
		Name:   RDSProxyResource,
		Scope:  nuke.Account,
		Lister: &RDSProxyLister{},
	})
}

type RDSProxyLister struct{}

func (l *RDSProxyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBProxiesInput{}
	resp, err := svc.DescribeDBProxies(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, instance := range resp.DBProxies {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: instance.DBProxyArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSProxy{
			svc:  svc,
			id:   *instance.DBProxyName,
			tags: tags.TagList,
		})
	}

	return resources, nil
}

type RDSProxy struct {
	svc  *rds.RDS
	id   string
	tags []*rds.Tag
}

func (i *RDSProxy) Remove(_ context.Context) error {
	params := &rds.DeleteDBProxyInput{
		DBProxyName: &i.id,
	}

	_, err := i.svc.DeleteDBProxy(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSProxy) String() string {
	return i.id
}

func (i *RDSProxy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ProxyName", i.id)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
