package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

const ResourceGroupGroupResource = "ResourceGroupGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   ResourceGroupGroupResource,
		Scope:  nuke.Account,
		Lister: &ResourceGroupGroupLister{},
	})
}

type ResourceGroupGroupLister struct{}

func (l *ResourceGroupGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := resourcegroups.New(opts.Session)
	var resources []resource.Resource

	params := &resourcegroups.ListGroupsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListGroups(params)
		if err != nil {
			return nil, err
		}

		for _, group := range output.Groups {
			resources = append(resources, &ResourceGroupGroup{
				svc:       svc,
				groupName: group.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ResourceGroupGroup struct {
	svc       *resourcegroups.ResourceGroups
	groupName *string
}

func (f *ResourceGroupGroup) Remove(_ context.Context) error {

	_, err := f.svc.DeleteGroup(&resourcegroups.DeleteGroupInput{
		Group: f.groupName,
	})

	return err
}

func (f *ResourceGroupGroup) String() string {
	return *f.groupName
}
