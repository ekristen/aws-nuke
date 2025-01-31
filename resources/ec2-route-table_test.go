package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2RouteTable_Filter(t *testing.T) {
	cases := []struct {
		ownerID   *string
		accountID *string
		filtered  bool
	}{
		{
			ownerID:   ptr.String("123456789012"),
			accountID: ptr.String("123456789012"),
			filtered:  false,
		},
		{
			ownerID:   ptr.String("123456789012"),
			accountID: ptr.String("123456789013"),
			filtered:  true,
		},
	}

	for _, c := range cases {
		r := EC2RouteTable{
			svc:        nil,
			accountID:  c.accountID,
			routeTable: &ec2.RouteTable{OwnerId: c.ownerID},
			vpc:        &ec2.Vpc{VpcId: ptr.String("vpc-12345678"), OwnerId: c.ownerID},
		}

		err := r.Filter()

		if c.filtered {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
