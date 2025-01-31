package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2DHCPOption_Filter(t *testing.T) {
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
		r := EC2DHCPOption{
			svc:       nil,
			accountID: c.accountID,
			ownerID:   c.ownerID,
		}

		err := r.Filter()

		if c.filtered {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
