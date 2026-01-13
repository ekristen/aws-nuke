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
			Name:         function.FunctionName,
			LastModified: function.LastModified,
			Tags:         tags.Tags,
		})
	}

	return resources, nil
}

type LambdaFunction struct {
	svc          *lambda.Lambda
	Name         *string
	LastModified *string
	Tags         map[string]*string
}

func (r *LambdaFunction) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *LambdaFunction) Remove(_ context.Context) error {
	_, err := r.svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: r.Name,
	})

	return err
}

func (r *LambdaFunction) String() string {
	return *r.Name
}
