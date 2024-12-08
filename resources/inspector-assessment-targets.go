package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/inspector"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const InspectorAssessmentTargetResource = "InspectorAssessmentTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:     InspectorAssessmentTargetResource,
		Scope:    nuke.Account,
		Resource: &InspectorAssessmentTarget{},
		Lister:   &InspectorAssessmentTargetLister{},
	})
}

type InspectorAssessmentTargetLister struct{}

func (l *InspectorAssessmentTargetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := inspector.New(opts.Session)

	resp, err := svc.ListAssessmentTargets(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.AssessmentTargetArns {
		resources = append(resources, &InspectorAssessmentTarget{
			svc: svc,
			arn: *out,
		})
	}

	return resources, nil
}

type InspectorAssessmentTarget struct {
	svc *inspector.Inspector
	arn string
}

func (e *InspectorAssessmentTarget) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAssessmentTarget(&inspector.DeleteAssessmentTargetInput{
		AssessmentTargetArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *InspectorAssessmentTarget) String() string {
	return e.arn
}
