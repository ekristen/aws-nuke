package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type GlueWorkflow struct {
	svc  *glue.Glue
	name *string
}

const GlueWorkflowResource = "GlueWorkflow"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueWorkflowResource,
		Scope:  nuke.Account,
		Lister: &GlueWorkflowLister{},
	})
}

type GlueWorkflowLister struct{}

func (l *GlueWorkflowLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.ListWorkflowsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		output, err := svc.ListWorkflows(params)
		if err != nil {
			return nil, err
		}

		for _, workflowName := range output.Workflows {
			resources = append(resources, &GlueWorkflow{
				svc:  svc,
				name: workflowName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *GlueWorkflow) Remove(_ context.Context) error {
	_, err := f.svc.DeleteWorkflow(&glue.DeleteWorkflowInput{
		Name: f.name,
	})

	return err
}

func (f *GlueWorkflow) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)

	return properties
}

func (f *GlueWorkflow) String() string {
	return *f.name
}
