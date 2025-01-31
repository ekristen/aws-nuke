package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func Test_EC2SecurityGroup_Properties(t *testing.T) {
	r := &EC2SecurityGroup{
		ID:      ptr.String("sg-12345678"),
		Name:    ptr.String("testing"),
		OwnerID: ptr.String("123456789012"),
		Tags: []*ec2.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("test"),
			},
		},
	}

	props := r.Properties()

	assert.Equal(t, ptr.ToString(r.ID), props.Get("ID"))
	assert.Equal(t, ptr.ToString(r.Name), props.Get("Name"))
	assert.Equal(t, ptr.ToString(r.OwnerID), props.Get("OwnerID"))
	assert.Equal(t, "test", props.Get("tag:Name"))

	assert.Equal(t, ptr.ToString(r.ID), r.String())
}

func Test_EC2SecurityGroup_Filter(t *testing.T) {
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
		r := &EC2SecurityGroup{
			accountID: c.accountID,
			OwnerID:   c.ownerID,
		}

		err := r.Filter()
		if c.filtered {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
