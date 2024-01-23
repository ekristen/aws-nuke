package resources

import (
	"context"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type DAXParameterGroup struct {
	svc                *dax.DAX
	parameterGroupName *string
}

const DAXParameterGroupResource = "DAXParameterGroup"

func init() {
	resource.Register(&resource.Registration{
		Name:   DAXParameterGroupResource,
		Scope:  nuke.Account,
		Lister: &DAXParameterGroupLister{},
	})
}

type DAXParameterGroupLister struct{}

func (l *DAXParameterGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := dax.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &dax.DescribeParameterGroupsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeParameterGroups(params)
		if err != nil {
			return nil, err
		}

		for _, parameterGroup := range output.ParameterGroups {
			//Ensure default is not deleted
			if !strings.Contains(*parameterGroup.ParameterGroupName, "default") {
				resources = append(resources, &DAXParameterGroup{
					svc:                svc,
					parameterGroupName: parameterGroup.ParameterGroupName,
				})
			}
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *DAXParameterGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteParameterGroup(&dax.DeleteParameterGroupInput{
		ParameterGroupName: f.parameterGroupName,
	})

	return err
}

func (f *DAXParameterGroup) String() string {
	return *f.parameterGroupName
}
