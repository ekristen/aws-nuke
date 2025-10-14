package resources

import (
	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck
	"github.com/aws/smithy-go/ptr"
)

func DefaultVpc(svc *ec2.EC2) *ec2.Vpc {
	resp, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   ptr.String("is-default"),
				Values: ptr.StringSlice([]string{"true"}),
			},
		},
	})
	if err != nil {
		return nil
	}

	if len(resp.Vpcs) == 0 {
		return nil
	}

	return resp.Vpcs[0]
}

func GetVPC(svc *ec2.EC2, vpcID *string) (*ec2.Vpc, error) {
	resp, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{vpcID},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Vpcs) == 0 {
		return nil, nil //nolint:nilnil
	}

	return resp.Vpcs[0], nil
}
