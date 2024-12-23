package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2TGWResource = "EC2TGW"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2TGWResource,
		Scope:    nuke.Account,
		Resource: &EC2TGW{},
		Lister:   &EC2TGWLister{},
		DependsOn: []string{
			EC2TGWAttachmentResource,
		},
	})
}

type EC2TGWLister struct{}

func (l *EC2TGWLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeTransitGatewaysInput{}
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.DescribeTransitGateways(params)
		if err != nil {
			return nil, err
		}

		for _, tgw := range resp.TransitGateways {
			resources = append(resources, &EC2TGW{
				svc:       svc,
				ID:        tgw.TransitGatewayId,
				OwnerID:   tgw.OwnerId,
				Tags:      tgw.Tags,
				State:     tgw.State,
				accountID: opts.AccountID,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &ec2.DescribeTransitGatewaysInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type EC2TGW struct {
	svc     *ec2.EC2
	ID      *string    `description:"The ID of the transit gateway."`
	OwnerID *string    `property:"name=OwnerId" description:"The ID of the AWS account that owns the transit gateway."`
	State   *string    `description:"The state of the transit gateway."`
	Tags    []*ec2.Tag `description:"The tags associated with the transit gateway."`

	accountID *string
}

func (r *EC2TGW) Remove(_ context.Context) error {
	params := &ec2.DeleteTransitGatewayInput{
		TransitGatewayId: r.ID,
	}

	_, err := r.svc.DeleteTransitGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2TGW) Filter() error {
	if ptr.ToString(r.State) == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}
	if ptr.ToString(r.OwnerID) != ptr.ToString(r.accountID) {
		return fmt.Errorf("not owned by account")
	}

	return nil
}

func (r *EC2TGW) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2TGW) String() string {
	return *r.ID
}
