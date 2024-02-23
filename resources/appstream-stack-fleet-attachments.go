package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppStreamStackFleetAttachment struct {
	svc       *appstream.AppStream
	stackName *string
	fleetName *string
}

const AppStreamStackFleetAttachmentResource = "AppStreamStackFleetAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppStreamStackFleetAttachmentResource,
		Scope:  nuke.Account,
		Lister: &AppStreamStackFleetAttachmentLister{},
	})
}

type AppStreamStackFleetAttachmentLister struct{}

func (l *AppStreamStackFleetAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var stacks []*appstream.Stack
	params := &appstream.DescribeStacksInput{}

	for {
		output, err := svc.DescribeStacks(params)
		if err != nil {
			return nil, err
		}

		stacks = append(stacks, output.Stacks...)

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	stackAssocParams := &appstream.ListAssociatedFleetsInput{}
	for _, stack := range stacks {
		stackAssocParams.StackName = stack.Name
		output, err := svc.ListAssociatedFleets(stackAssocParams)
		if err != nil {
			return nil, err
		}

		for _, name := range output.Names {
			resources = append(resources, &AppStreamStackFleetAttachment{
				svc:       svc,
				stackName: stack.Name,
				fleetName: name,
			})
		}
	}

	return resources, nil
}

func (f *AppStreamStackFleetAttachment) Remove(_ context.Context) error {
	_, err := f.svc.DisassociateFleet(&appstream.DisassociateFleetInput{
		StackName: f.stackName,
		FleetName: f.fleetName,
	})

	return err
}

func (f *AppStreamStackFleetAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.stackName, *f.fleetName)
}
