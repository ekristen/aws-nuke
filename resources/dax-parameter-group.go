package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/dax" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DAXParameterGroupResource = "DAXParameterGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     DAXParameterGroupResource,
		Scope:    nuke.Account,
		Resource: &DAXParameterGroup{},
		Lister:   &DAXParameterGroupLister{},
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
			resources = append(resources, &DAXParameterGroup{
				svc:  svc,
				Name: parameterGroup.ParameterGroupName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type DAXParameterGroup struct {
	svc  *dax.DAX
	Name *string
}

func (r *DAXParameterGroup) Filter() error {
	if strings.Contains(*r.Name, "default") { //nolint:goconst,nolintlint
		return fmt.Errorf("unable to delete default")
	}

	return nil
}

func (r *DAXParameterGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteParameterGroup(&dax.DeleteParameterGroupInput{
		ParameterGroupName: r.Name,
	})

	return err
}

func (r *DAXParameterGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DAXParameterGroup) String() string {
	return *r.Name
}
