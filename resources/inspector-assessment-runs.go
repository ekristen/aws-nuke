package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/inspector" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const InspectorAssessmentRunResource = "InspectorAssessmentRun"

func init() {
	registry.Register(&registry.Registration{
		Name:     InspectorAssessmentRunResource,
		Scope:    nuke.Account,
		Resource: &InspectorAssessmentRun{},
		Lister:   &InspectorAssessmentRunLister{},
	})
}

type InspectorAssessmentRunLister struct{}

func (l *InspectorAssessmentRunLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := inspector.New(opts.Session)

	resp, err := svc.ListAssessmentRuns(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.AssessmentRunArns {
		resources = append(resources, &InspectorAssessmentRun{
			svc: svc,
			arn: *out,
		})
	}

	return resources, nil
}

type InspectorAssessmentRun struct {
	svc *inspector.Inspector
	arn string
}

func (e *InspectorAssessmentRun) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAssessmentRun(&inspector.DeleteAssessmentRunInput{
		AssessmentRunArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *InspectorAssessmentRun) String() string {
	return e.arn
}
