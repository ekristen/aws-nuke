package resources

import (
	"context"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RedshiftParameterGroupResource = "RedshiftParameterGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   RedshiftParameterGroupResource,
		Scope:  nuke.Account,
		Lister: &RedshiftParameterGroupLister{},
	})
}

type RedshiftParameterGroupLister struct{}

func (l *RedshiftParameterGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshift.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshift.DescribeClusterParameterGroupsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeClusterParameterGroups(params)
		if err != nil {
			return nil, err
		}

		for _, parameterGroup := range output.ParameterGroups {
			if !strings.Contains(*parameterGroup.ParameterGroupName, "default.redshift") {
				resources = append(resources, &RedshiftParameterGroup{
					svc:                svc,
					parameterGroupName: parameterGroup.ParameterGroupName,
				})
			}
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type RedshiftParameterGroup struct {
	svc                *redshift.Redshift
	parameterGroupName *string
}

func (f *RedshiftParameterGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteClusterParameterGroup(&redshift.DeleteClusterParameterGroupInput{
		ParameterGroupName: f.parameterGroupName,
	})

	return err
}

func (f *RedshiftParameterGroup) String() string {
	return *f.parameterGroupName
}
