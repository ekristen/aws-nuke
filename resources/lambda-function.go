package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LambdaFunctionResource = "LambdaFunction"

// lambdaLastModifiedLayout is the format of FunctionConfiguration.LastModified, for example
// 2019-08-14T22:32:35.183+0000. It is not RFC 3339 (no colon in the offset).
const lambdaLastModifiedLayout = "2006-01-02T15:04:05.000-0700"

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

		for i := range resp.Functions {
			function := &resp.Functions[i]
			tags, err := svc.ListTags(ctx, &lambda.ListTagsInput{
				Resource: function.FunctionArn,
			})
			if err != nil {
				continue
			}

			resources = append(resources, &LambdaFunction{
				svc:          svc,
				Name:         function.FunctionName,
				LastModified: parseLambdaLastModified(function.LastModified),
				Tags:         tags.Tags,
			})
		}
	}

	return resources, nil
}

type LambdaFunction struct {
	svc          *lambda.Client
	Name         *string
	LastModified *time.Time
	Tags         map[string]string
}

func parseLambdaLastModified(value *string) *time.Time {
	if value == nil {
		return nil
	}

	t, err := time.Parse(lambdaLastModifiedLayout, *value)
	if err != nil {
		return nil
	}

	t = t.UTC()
	return &t
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
