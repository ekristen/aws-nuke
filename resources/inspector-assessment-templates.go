package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/inspector"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const InspectorAssessmentTemplateResource = "InspectorAssessmentTemplate"

func init() {
	registry.Register(&registry.Registration{
		Name:     InspectorAssessmentTemplateResource,
		Scope:    nuke.Account,
		Resource: &InspectorAssessmentTemplate{},
		Lister:   &InspectorAssessmentTemplateLister{},
	})
}

type InspectorAssessmentTemplateLister struct{}

func (l *InspectorAssessmentTemplateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := inspector.New(opts.Session)

	resp, err := svc.ListAssessmentTemplates(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.AssessmentTemplateArns {
		resources = append(resources, &InspectorAssessmentTemplate{
			svc: svc,
			arn: *out,
		})
	}

	return resources, nil
}

type InspectorAssessmentTemplate struct {
	svc *inspector.Inspector
	arn string
}

func (e *InspectorAssessmentTemplate) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAssessmentTemplate(&inspector.DeleteAssessmentTemplateInput{
		AssessmentTemplateArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *InspectorAssessmentTemplate) String() string {
	return e.arn
}
