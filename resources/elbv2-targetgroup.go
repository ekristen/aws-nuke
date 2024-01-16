package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ELBv2TargetGroupResource = "ELBv2TargetGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   ELBv2TargetGroupResource,
		Scope:  nuke.Account,
		Lister: &ELBv2TargetGroupLister{},
		DependsOn: []string{
			ELBv2Resource,
		},
	})
}

type ELBv2TargetGroupLister struct{}

func (l *ELBv2TargetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elbv2.New(opts.Session)
	var tagReqELBv2TargetGroupARNs []*string
	targetGroupARNToRsc := make(map[string]*elbv2.TargetGroup)

	err := svc.DescribeTargetGroupsPages(nil,
		func(page *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
			for _, targetGroup := range page.TargetGroups {
				tagReqELBv2TargetGroupARNs = append(tagReqELBv2TargetGroupARNs, targetGroup.TargetGroupArn)
				targetGroupARNToRsc[*targetGroup.TargetGroupArn] = targetGroup
			}
			return !lastPage
		})
	if err != nil {
		return nil, err
	}

	// Tags for ELBv2 target groups need to be fetched separately
	// We can only specify up to 20 in a single call
	// See: https://github.com/aws/aws-sdk-go/blob/0e8c61841163762f870f6976775800ded4a789b0/service/elbv2/api.go#L5398
	resources := make([]resource.Resource, 0)
	for len(tagReqELBv2TargetGroupARNs) > 0 {
		requestElements := len(tagReqELBv2TargetGroupARNs)
		if requestElements > 20 {
			requestElements = 20
		}

		tagResp, err := svc.DescribeTags(&elbv2.DescribeTagsInput{
			ResourceArns: tagReqELBv2TargetGroupARNs[:requestElements],
		})
		if err != nil {
			return nil, err
		}
		for _, tagInfo := range tagResp.TagDescriptions {
			resources = append(resources, &ELBv2TargetGroup{
				svc:  svc,
				tg:   targetGroupARNToRsc[*tagInfo.ResourceArn],
				tags: tagInfo.Tags,
			})
		}

		// Remove the elements that were queried
		tagReqELBv2TargetGroupARNs = tagReqELBv2TargetGroupARNs[requestElements:]
	}
	return resources, nil
}

type ELBv2TargetGroup struct {
	svc  *elbv2.ELBV2
	tg   *elbv2.TargetGroup
	tags []*elbv2.Tag
}

func (e *ELBv2TargetGroup) Remove(_ context.Context) error {
	_, err := e.svc.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
		TargetGroupArn: e.tg.TargetGroupArn,
	})

	if err != nil {
		return err
	}

	return nil
}

func (e *ELBv2TargetGroup) Properties() types.Properties {
	properties := types.NewProperties().
		Set("ARN", e.tg.TargetGroupArn)

	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	properties.Set("IsLoadBalanced", len(e.tg.LoadBalancerArns) > 0)

	return properties
}

func (e *ELBv2TargetGroup) String() string {
	return *e.tg.TargetGroupName
}
