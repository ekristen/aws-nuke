package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/codeguruprofiler" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeGuruProfilingGroupResource = "CodeGuruProfilingGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeGuruProfilingGroupResource,
		Scope:    nuke.Account,
		Resource: &CodeGuruProfilingGroup{},
		Lister:   &CodeGuruProfilingGroupResourceLister{},
	})
}

type CodeGuruProfilingGroupResourceLister struct{}

func (l *CodeGuruProfilingGroupResourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := codeguruprofiler.New(opts.Session)

	params := &codeguruprofiler.ListProfilingGroupsInput{
		IncludeDescription: aws.Bool(true),
	}

	for {
		resp, err := svc.ListProfilingGroups(params)
		if err != nil {
			return nil, err
		}

		for _, group := range resp.ProfilingGroups {
			resources = append(resources, &CodeGuruProfilingGroup{
				svc:             svc,
				ComputePlatform: group.ComputePlatform,
				Name:            group.Name,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeGuruProfilingGroup struct {
	svc             *codeguruprofiler.CodeGuruProfiler
	ComputePlatform *string
	Name            *string
}

func (r *CodeGuruProfilingGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteProfilingGroup(&codeguruprofiler.DeleteProfilingGroupInput{
		ProfilingGroupName: r.Name,
	})
	return err
}

func (r *CodeGuruProfilingGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CodeGuruProfilingGroup) String() string {
	return *r.Name
}
