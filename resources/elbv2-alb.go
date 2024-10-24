package resources

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type ELBv2LoadBalancer struct {
	svc      *elbv2.ELBV2
	tags     []*elbv2.Tag
	elb      *elbv2.LoadBalancer
	settings *libsettings.Setting
}

const ELBv2Resource = "ELBv2"

func init() {
	registry.Register(&registry.Registration{
		Name:   ELBv2Resource,
		Scope:  nuke.Account,
		Lister: &ELBv2Lister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type ELBv2Lister struct{}

func (l *ELBv2Lister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elbv2.New(opts.Session)
	var tagReqELBv2ARNs []*string
	elbv2ARNToRsc := make(map[string]*elbv2.LoadBalancer)

	err := svc.DescribeLoadBalancersPages(nil,
		func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
			for _, elbv2 := range page.LoadBalancers {
				tagReqELBv2ARNs = append(tagReqELBv2ARNs, elbv2.LoadBalancerArn)
				elbv2ARNToRsc[*elbv2.LoadBalancerArn] = elbv2
			}
			return !lastPage
		})

	if err != nil {
		return nil, err
	}

	// Tags for ELBv2s need to be fetched separately
	// We can only specify up to 20 in a single call
	// See: https://github.com/aws/aws-sdk-go/blob/0e8c61841163762f870f6976775800ded4a789b0/service/elbv2/api.go#L5398
	resources := make([]resource.Resource, 0)
	for len(tagReqELBv2ARNs) > 0 {
		requestElements := len(tagReqELBv2ARNs)
		if requestElements > 20 {
			requestElements = 20
		}

		tagResp, err := svc.DescribeTags(&elbv2.DescribeTagsInput{
			ResourceArns: tagReqELBv2ARNs[:requestElements],
		})
		if err != nil {
			return nil, err
		}
		for _, elbv2TagInfo := range tagResp.TagDescriptions {
			elb := elbv2ARNToRsc[*elbv2TagInfo.ResourceArn]
			resources = append(resources, &ELBv2LoadBalancer{
				svc:  svc,
				elb:  elb,
				tags: elbv2TagInfo.Tags,
			})
		}

		// Remove the elements that were queried
		tagReqELBv2ARNs = tagReqELBv2ARNs[requestElements:]
	}
	return resources, nil
}

func (e *ELBv2LoadBalancer) Settings(setting *libsettings.Setting) {
	e.settings = setting
}

func (e *ELBv2LoadBalancer) Remove(_ context.Context) error {
	params := &elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: e.elb.LoadBalancerArn,
	}

	if _, err := e.svc.DeleteLoadBalancer(params); err != nil {
		if e.settings.GetBool("DisableDeletionProtection") {
			var awsErr awserr.Error
			ok := errors.As(err, &awsErr)
			if ok && awsErr.Code() == "OperationNotPermitted" &&
				awsErr.Message() == "Load balancer '"+*e.elb.LoadBalancerArn+"' cannot be deleted because deletion protection is enabled" {
				err = e.DisableProtection()
				if err != nil {
					return err
				}

				_, err := e.svc.DeleteLoadBalancer(params)
				if err != nil {
					return err
				}

				return nil
			}
		}

		return err
	}

	return nil
}

func (e *ELBv2LoadBalancer) DisableProtection() error {
	params := &elbv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: e.elb.LoadBalancerArn,
		Attributes: []*elbv2.LoadBalancerAttribute{
			{
				Key:   aws.String("deletion_protection.enabled"),
				Value: aws.String("false"),
			},
		},
	}

	_, err := e.svc.ModifyLoadBalancerAttributes(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *ELBv2LoadBalancer) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreatedTime", e.elb.CreatedTime.Format(time.RFC3339)).
		Set("ARN", e.elb.LoadBalancerArn)

	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *ELBv2LoadBalancer) String() string {
	return *e.elb.LoadBalancerName
}
