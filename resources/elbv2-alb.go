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

const ELBv2Resource = "ELBv2"

func init() {
	registry.Register(&registry.Registration{
		Name:     ELBv2Resource,
		Scope:    nuke.Account,
		Resource: &ELBv2LoadBalancer{},
		Lister:   &ELBv2Lister{},
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
			for _, elbv2lb := range page.LoadBalancers {
				tagReqELBv2ARNs = append(tagReqELBv2ARNs, elbv2lb.LoadBalancerArn)
				elbv2ARNToRsc[*elbv2lb.LoadBalancerArn] = elbv2lb
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
				svc:         svc,
				ARN:         elb.LoadBalancerArn,
				Name:        elb.LoadBalancerName,
				CreatedTime: elb.CreatedTime,
				Tags:        elbv2TagInfo.Tags,
			})
		}

		// Remove the elements that were queried
		tagReqELBv2ARNs = tagReqELBv2ARNs[requestElements:]
	}
	return resources, nil
}

type ELBv2LoadBalancer struct {
	svc         *elbv2.ELBV2
	settings    *libsettings.Setting
	ARN         *string    `description:"ARN of the load balancer"`
	Name        *string    `description:"Name of the load balancer"`
	CreatedTime *time.Time `description:"Creation time of the load balancer"`
	Tags        []*elbv2.Tag
}

func (r *ELBv2LoadBalancer) Settings(setting *libsettings.Setting) {
	r.settings = setting
}

func (r *ELBv2LoadBalancer) Remove(_ context.Context) error {
	params := &elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: r.ARN,
	}

	if _, err := r.svc.DeleteLoadBalancer(params); err != nil {
		if r.settings.GetBool("DisableDeletionProtection") {
			var awsErr awserr.Error
			ok := errors.As(err, &awsErr)
			if ok && awsErr.Code() == "OperationNotPermitted" &&
				awsErr.Message() == "Load balancer '"+*r.ARN+"' cannot be deleted because deletion protection is enabled" {
				err = r.DisableProtection()
				if err != nil {
					return err
				}

				_, err := r.svc.DeleteLoadBalancer(params)
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

func (r *ELBv2LoadBalancer) DisableProtection() error {
	params := &elbv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: r.ARN,
		Attributes: []*elbv2.LoadBalancerAttribute{
			{
				Key:   aws.String("deletion_protection.enabled"),
				Value: aws.String("false"),
			},
		},
	}

	_, err := r.svc.ModifyLoadBalancerAttributes(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *ELBv2LoadBalancer) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ELBv2LoadBalancer) String() string {
	return *r.Name
}
