package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AutoScalingPlansScalingPlanResource = "AutoScalingPlansScalingPlan"

func init() {
	resource.Register(resource.Registration{
		Name:   AutoScalingPlansScalingPlanResource,
		Scope:  nuke.Account,
		Lister: &AutoScalingPlansScalingPlanLister{},
	})
}

type AutoScalingPlansScalingPlanLister struct{}

func (l *AutoScalingPlansScalingPlanLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := autoscalingplans.New(opts.Session)
	svc.ClientInfo.SigningName = "autoscaling-plans"
	resources := make([]resource.Resource, 0)

	params := &autoscalingplans.DescribeScalingPlansInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeScalingPlans(params)
		if err != nil {
			return nil, err
		}

		for _, scalingPlan := range output.ScalingPlans {
			resources = append(resources, &AutoScalingPlansScalingPlan{
				svc:                svc,
				scalingPlanName:    scalingPlan.ScalingPlanName,
				scalingPlanVersion: scalingPlan.ScalingPlanVersion,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type AutoScalingPlansScalingPlan struct {
	svc                *autoscalingplans.AutoScalingPlans
	scalingPlanName    *string
	scalingPlanVersion *int64
}

func (f *AutoScalingPlansScalingPlan) Remove(_ context.Context) error {
	_, err := f.svc.DeleteScalingPlan(&autoscalingplans.DeleteScalingPlanInput{
		ScalingPlanName:    f.scalingPlanName,
		ScalingPlanVersion: f.scalingPlanVersion,
	})

	return err
}

func (f *AutoScalingPlansScalingPlan) String() string {
	return *f.scalingPlanName
}
