package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"               //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/appstream" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppStreamImageBuilderWaiterResource = "AppStreamImageBuilderWaiter"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppStreamImageBuilderWaiterResource,
		Scope:    nuke.Account,
		Resource: &AppStreamImageBuilderWaiter{},
		Lister:   &AppStreamImageBuilderWaiterLister{},
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

type AppStreamImageBuilderWaiter struct {
	svc   *appstream.AppStream
	name  *string
	state *string
}

func (f *AppStreamImageBuilderWaiter) Remove(_ context.Context) error {
	return nil
}

func (f *AppStreamImageBuilderWaiter) String() string {
	return *f.name
}

func (f *AppStreamImageBuilderWaiter) Filter() error {
	if ptr.ToString(f.state) == appstream.ImageBuilderStateStopped {
		return fmt.Errorf("already stopped")
	} else if ptr.ToString(f.state) == appstream.ImageBuilderStateRunning {
		return fmt.Errorf("already running")
	} else if ptr.ToString(f.state) == appstream.ImageBuilderStateDeleting {
		return fmt.Errorf("already being deleted")
	}

	return nil
}
