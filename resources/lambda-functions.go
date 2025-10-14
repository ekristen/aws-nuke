package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lambda" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LambdaFunctionResource = "LambdaFunction"

func init() {
	registry.Register(&registry.Registration{
		Name:     LambdaFunctionResource,
		Scope:    nuke.Account,
		Resource: &LambdaFunction{},
		Lister:   &LambdaFunctionLister{},
	})
}

type LambdaFunctionLister struct{}

func (l *LambdaFunctionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lambda.New(opts.Session)

	functions := make([]*lambda.FunctionConfiguration, 0)

	params := &lambda.ListFunctionsInput{}

	err := svc.ListFunctionsPages(params, func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
		functions = append(functions, page.Functions...)
		return true
	})

	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, function := range functions {
		tags, err := svc.ListTags(&lambda.ListTagsInput{
			Resource: function.FunctionArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &LambdaFunction{
			svc:          svc,
			functionName: function.FunctionName,
			tags:         tags.Tags,
		})
	}

	return resources, nil
}

type LambdaFunction struct {
	svc          *lambda.Lambda
	functionName *string
	tags         map[string]*string
}

func (f *LambdaFunction) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.functionName)

	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}

	return properties
}

func (f *LambdaFunction) Remove(_ context.Context) error {
	_, err := f.svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: f.functionName,
	})

	return err
}

func (f *LambdaFunction) String() string {
	return *f.functionName
}
