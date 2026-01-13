package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"

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

func (l *LambdaFunctionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := lambda.NewFromConfig(*opts.Config)

	resources := make([]resource.Resource, 0)

	params := &lambda.ListFunctionsInput{}

	paginator := lambda.NewListFunctionsPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, function := range resp.Functions {
			tags, err := svc.ListTags(ctx, &lambda.ListTagsInput{
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
	}

	return resources, nil
}

type LambdaFunction struct {
	svc          *lambda.Client
	Name         *string
	LastModified *string
	Tags         map[string]string
}

func (r *LambdaFunction) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *LambdaFunction) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: r.Name,
	})

	return err
}

func (r *LambdaFunction) String() string {
	return *r.Name
}
