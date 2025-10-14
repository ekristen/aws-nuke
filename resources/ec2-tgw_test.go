package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck
)

func Test_EC2TGW_Properties(t *testing.T) {
	tgw := &EC2TGW{
		ID:      ptr.String("tgw-1234567890abcdef0"),
		OwnerID: ptr.String("123456789012"),
		Tags: []*ec2.Tag{
			{
				Key:   ptr.String("TestTag"),
				Value: ptr.String("test-tgw"),
			},
		},
	}

	assert.Equal(t, "tgw-1234567890abcdef0", tgw.Properties().Get("ID"))
	assert.Equal(t, "123456789012", tgw.Properties().Get("OwnerId"))
	assert.Equal(t, "test-tgw", tgw.Properties().Get("tag:TestTag"))
	assert.Equal(t, "tgw-1234567890abcdef0", tgw.String())
}
