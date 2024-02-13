package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/applicationautoscaling"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ApplicationAutoScalingScalableTargetResource = "ApplicationAutoScalingScalableTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:   ApplicationAutoScalingScalableTargetResource,
		Scope:  nuke.Account,
		Lister: &ApplicationAutoScalingScalableTargetLister{},
	})
}

type ApplicationAutoScalingScalableTargetLister struct{}

func (l *ApplicationAutoScalingScalableTargetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := applicationautoscaling.New(opts.Session)

	namespaces := applicationautoscaling.ServiceNamespace_Values()

	params := &applicationautoscaling.DescribeScalableTargetsInput{}
	resources := make([]resource.Resource, 0)
	for _, namespace := range namespaces {
		for {
			params.ServiceNamespace = &namespace
			resp, err := svc.DescribeScalableTargets(params)
			if err != nil {
				return nil, err
			}

			for _, out := range resp.ScalableTargets {
				resources = append(resources, &AppAutoScaling{
					svc:       svc,
					target:    out,
					id:        *out.ResourceId,
					roleARN:   *out.RoleARN,
					dimension: *out.ScalableDimension,
					namespace: *out.ServiceNamespace,
				})
			}

			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type AppAutoScaling struct {
	svc       *applicationautoscaling.ApplicationAutoScaling
	target    *applicationautoscaling.ScalableTarget
	id        string
	roleARN   string
	dimension string
	namespace string
}

func (a *AppAutoScaling) Remove(_ context.Context) error {
	_, err := a.svc.DeregisterScalableTarget(&applicationautoscaling.DeregisterScalableTargetInput{
		ResourceId:        &a.id,
		ScalableDimension: &a.dimension,
		ServiceNamespace:  &a.namespace,
	})

	if err != nil {
		return err
	}

	return nil
}

func (a *AppAutoScaling) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ResourceID", a.id)
	properties.Set("ScalableDimension", a.dimension)
	properties.Set("ServiceNamespace", a.namespace)

	return properties
}

func (a *AppAutoScaling) String() string {
	return a.id + ": " + a.dimension
}
