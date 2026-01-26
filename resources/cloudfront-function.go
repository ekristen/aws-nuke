package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudfront" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFrontFunctionResource = "CloudFrontFunction"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFrontFunctionResource,
		Scope:    nuke.Account,
		Resource: &CloudFrontFunction{},
		Lister:   &CloudFrontFunctionLister{},
	})
}

type CloudFrontFunctionLister struct{}

func (l *CloudFrontFunctionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListFunctionsInput{}

	for {
		resp, err := svc.ListFunctions(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.FunctionList.Items {
			resources = append(resources, &CloudFrontFunction{
				svc:   svc,
				name:  item.Name,
				stage: item.FunctionMetadata.Stage,
			})
		}

		if resp.FunctionList.NextMarker == nil {
			break
		}

		params.Marker = resp.FunctionList.NextMarker
	}

	return resources, nil
}

type CloudFrontFunction struct {
	svc   *cloudfront.CloudFront
	name  *string
	stage *string
}

func (f *CloudFrontFunction) Remove(_ context.Context) error {
	resp, err := f.svc.GetFunction(&cloudfront.GetFunctionInput{
		Name:  f.name,
		Stage: f.stage,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteFunction(&cloudfront.DeleteFunctionInput{
		Name:    f.name,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontFunction) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("name", f.name)
	properties.Set("stage", f.stage)
	return properties
}

func (f *CloudFrontFunction) String() string {
	return *f.name
}
