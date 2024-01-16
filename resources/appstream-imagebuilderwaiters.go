package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppStreamImageBuilderWaiter struct {
	svc   *appstream.AppStream
	name  *string
	state *string
}

const AppStreamImageBuilderWaiterResource = "AppStreamImageBuilderWaiter"

func init() {
	resource.Register(resource.Registration{
		Name:   AppStreamImageBuilderWaiterResource,
		Scope:  nuke.Account,
		Lister: &AppStreamImageBuilderWaiterLister{},
	})
}

type AppStreamImageBuilderWaiterLister struct{}

func (l *AppStreamImageBuilderWaiterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &appstream.DescribeImageBuildersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeImageBuilders(params)
		if err != nil {
			return nil, err
		}

		for _, imageBuilder := range output.ImageBuilders {
			resources = append(resources, &AppStreamImageBuilderWaiter{
				svc:   svc,
				name:  imageBuilder.Name,
				state: imageBuilder.State,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamImageBuilderWaiter) Remove(_ context.Context) error {

	return nil
}

func (f *AppStreamImageBuilderWaiter) String() string {
	return *f.name
}

func (f *AppStreamImageBuilderWaiter) Filter() error {
	if *f.state == "STOPPED" {
		return fmt.Errorf("already stopped")
	} else if *f.state == "RUNNING" {
		return fmt.Errorf("already running")
	} else if *f.state == "DELETING" {
		return fmt.Errorf("already being deleted")
	}

	return nil
}
