package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2InternetGatewayAttachment_Filter(t *testing.T) {
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
		r := EC2InternetGatewayAttachment{
			svc:        nil,
			accountID:  c.accountID,
			igwOwnerID: c.ownerID,
		}

		err := r.Filter()

		if c.filtered {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
