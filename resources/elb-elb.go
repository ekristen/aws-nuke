package resources

import (
	"context"

	"time"

	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ELBResource = "ELB"

func init() {
	registry.Register(&registry.Registration{
		Name:   ELBResource,
		Scope:  nuke.Account,
		Lister: &ELBLister{},
	})
}

type ELBLister struct{}

func (l *ELBLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	resources := make([]resource.Resource, 0)
	elbNames := make([]*string, 0)
	elbNameToRsc := make(map[string]*elb.LoadBalancerDescription)
	svc := elb.New(opts.Session)

	err := svc.DescribeLoadBalancersPages(nil,
		func(page *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
			for _, desc := range page.LoadBalancerDescriptions {
				elbNames = append(elbNames, desc.LoadBalancerName)
				elbNameToRsc[*desc.LoadBalancerName] = desc
			}
			return !lastPage
		})

	if err != nil {
		return nil, err
	}

	for len(elbNames) > 0 {
		requestElements := len(elbNames)
		if requestElements > 20 {
			requestElements = 20
		}

		tagResp, err := svc.DescribeTags(&elb.DescribeTagsInput{
			LoadBalancerNames: elbNames[:requestElements],
		})
		if err != nil {
			return nil, err
		}
		for _, elbTagInfo := range tagResp.TagDescriptions {
			elbEntity := elbNameToRsc[*elbTagInfo.LoadBalancerName]
			resources = append(resources, &ELBLoadBalancer{
				svc:  svc,
				elb:  elbEntity,
				tags: elbTagInfo.Tags,
			})
		}

		// Remove the elements that were queried
		elbNames = elbNames[requestElements:]
	}

	return resources, nil
}

type ELBLoadBalancer struct {
	svc  *elb.ELB
	elb  *elb.LoadBalancerDescription
	tags []*elb.Tag
}

func (e *ELBLoadBalancer) Remove(_ context.Context) error {
	params := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: e.elb.LoadBalancerName,
	}

	_, err := e.svc.DeleteLoadBalancer(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *ELBLoadBalancer) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreatedTime", e.elb.CreatedTime.Format(time.RFC3339))

	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *ELBLoadBalancer) String() string {
	return *e.elb.LoadBalancerName
}
