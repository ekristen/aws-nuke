package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Route53HealthCheckResource = "Route53HealthCheck"

func init() {
	registry.Register(&registry.Registration{
		Name:     Route53HealthCheckResource,
		Scope:    nuke.Account,
		Resource: &Route53HealthCheck{},
		Lister:   &Route53HealthCheckLister{},
	})
}

type Route53HealthCheckLister struct{}

func (l *Route53HealthCheckLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := route53.New(opts.Session)
	params := &route53.ListHealthChecksInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListHealthChecks(params)
		if err != nil {
			return nil, err
		}

		for _, check := range resp.HealthChecks {
			tags, err := svc.ListTagsForResource(&route53.ListTagsForResourceInput{
				ResourceId:   check.Id,
				ResourceType: aws.String("healthcheck"),
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &Route53HealthCheck{
				svc:  svc,
				id:   check.Id,
				tags: tags.ResourceTagSet.Tags,
			})
		}

		if !aws.BoolValue(resp.IsTruncated) {
			break
		}

		params.Marker = resp.NextMarker
	}

	return resources, nil
}

type Route53HealthCheck struct {
	svc  *route53.Route53
	id   *string
	tags []*route53.Tag
}

func (hz *Route53HealthCheck) Remove(_ context.Context) error {
	params := &route53.DeleteHealthCheckInput{
		HealthCheckId: hz.id,
	}

	_, err := hz.svc.DeleteHealthCheck(params)
	if err != nil {
		return err
	}

	return nil
}

func (hz *Route53HealthCheck) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", hz.id)

	for _, tag := range hz.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (hz *Route53HealthCheck) String() string {
	return ptr.ToString(hz.id)
}
