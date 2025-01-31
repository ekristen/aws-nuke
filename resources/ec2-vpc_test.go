package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VPC_Filter(t *testing.T) {
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
		r := EC2VPC{
			svc:       nil,
			vpc:       &ec2.Vpc{OwnerId: c.ownerID},
			accountID: c.accountID,
		}

		err := r.Filter()

		if c.filtered {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
