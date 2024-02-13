package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

type AppStreamStack struct {
	svc  *appstream.AppStream
	name *string
}

const AppStreamStackResource = "AppStreamStack"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppStreamStackResource,
		Scope:  nuke.Account,
		Lister: &AppStreamStackLister{},
	})
}

type AppStreamStackLister struct{}

func (l *AppStreamStackLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &appstream.DescribeStacksInput{}

	for {
		output, err := svc.DescribeStacks(params)
		if err != nil {
			return nil, err
		}

		for _, stack := range output.Stacks {
			resources = append(resources, &AppStreamStack{
				svc:  svc,
				name: stack.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamStack) Remove(_ context.Context) error {
	_, err := f.svc.DeleteStack(&appstream.DeleteStackInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamStack) String() string {
	return *f.name
}
