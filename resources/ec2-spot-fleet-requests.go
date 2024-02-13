package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2SpotFleetRequestResource = "EC2SpotFleetRequest"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2SpotFleetRequestResource,
		Scope:  nuke.Account,
		Lister: &EC2SpotFleetRequestLister{},
	})
}

type EC2SpotFleetRequestLister struct{}

func (l *EC2SpotFleetRequestLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeSpotFleetRequests(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, config := range resp.SpotFleetRequestConfigs {
		resources = append(resources, &EC2SpotFleetRequest{
			svc:   svc,
			id:    *config.SpotFleetRequestId,
			state: *config.SpotFleetRequestState,
		})
	}

	return resources, nil
}

type EC2SpotFleetRequest struct {
	svc   *ec2.EC2
	id    string
	state string
}

func (i *EC2SpotFleetRequest) Filter() error {
	if i.state == "cancelled" {
		return fmt.Errorf("already cancelled")
	}
	return nil
}

func (i *EC2SpotFleetRequest) Remove(_ context.Context) error {
	params := &ec2.CancelSpotFleetRequestsInput{
		TerminateInstances: aws.Bool(true),
		SpotFleetRequestIds: []*string{
			&i.id,
		},
	}

	_, err := i.svc.CancelSpotFleetRequests(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *EC2SpotFleetRequest) String() string {
	return i.id
}
