package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DAXSubnetGroupResource = "DAXSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   DAXSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &DAXSubnetGroupLister{},
	})
}

type DAXSubnetGroupLister struct{}

func (l *DAXSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := dax.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &dax.DescribeSubnetGroupsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeSubnetGroups(params)
		if err != nil {
			return nil, err
		}

		for _, subnet := range output.SubnetGroups {
			resources = append(resources, &DAXSubnetGroup{
				svc:             svc,
				subnetGroupName: subnet.SubnetGroupName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type DAXSubnetGroup struct {
	svc             *dax.DAX
	subnetGroupName *string
}

func (f *DAXSubnetGroup) Filter() error {
	if *f.subnetGroupName == "default" {
		return fmt.Errorf("cannot delete default DAX Subnet group")
	}
	return nil
}

func (f *DAXSubnetGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSubnetGroup(&dax.DeleteSubnetGroupInput{
		SubnetGroupName: f.subnetGroupName,
	})

	return err
}

func (f *DAXSubnetGroup) String() string {
	return *f.subnetGroupName
}
