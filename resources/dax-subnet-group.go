package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DAXSubnetGroupResource = "DAXSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     DAXSubnetGroupResource,
		Scope:    nuke.Account,
		Resource: &DAXSubnetGroup{},
		Lister:   &DAXSubnetGroupLister{},
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
				svc:  svc,
				Name: subnet.SubnetGroupName,
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
	svc  *dax.DAX
	Name *string
}

func (r *DAXSubnetGroup) Filter() error {
	if *r.Name == "default" { //nolint:goconst,nolintlint
		return fmt.Errorf("cannot delete default DAX Subnet group")
	}
	return nil
}

func (r *DAXSubnetGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSubnetGroup(&dax.DeleteSubnetGroupInput{
		SubnetGroupName: r.Name,
	})

	return err
}

func (r *DAXSubnetGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DAXSubnetGroup) String() string {
	return *r.Name
}
