package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2PlacementGroupResource = "EC2PlacementGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2PlacementGroupResource,
		Scope:  nuke.Account,
		Lister: &EC2PlacementGroupLister{},
	})
}

type EC2PlacementGroupLister struct{}

func (l *EC2PlacementGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribePlacementGroupsInput{}
	resp, err := svc.DescribePlacementGroups(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.PlacementGroups {
		resources = append(resources, &EC2PlacementGroup{
			svc:   svc,
			name:  *out.GroupName,
			state: *out.State,
		})
	}

	return resources, nil
}

type EC2PlacementGroup struct {
	svc   *ec2.EC2
	name  string
	state string
}

func (p *EC2PlacementGroup) Filter() error {
	if p.state == "deleted" {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (p *EC2PlacementGroup) Remove(_ context.Context) error {
	params := &ec2.DeletePlacementGroupInput{
		GroupName: &p.name,
	}

	_, err := p.svc.DeletePlacementGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (p *EC2PlacementGroup) String() string {
	return p.name
}
