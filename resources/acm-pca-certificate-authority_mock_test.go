package resources

import (
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Mock_ACMPCACertificateAuthority_Properties(t *testing.T) {
	r := &ACMPCACertificateAuthority{
		ARN:    ptr.String("certificateAuthorityARN"),
		Status: ptr.String("status"),
		Tags: []*acmpca.Tag{
			{
				Key:   ptr.String("key"),
				Value: ptr.String("value"),
			},
		},
	}

	properties := r.Properties()
	assert.Equal(t, "value", properties.Get("tag:key"))
	assert.Equal(t, "status", properties.Get("Status"))
	assert.Equal(t, "certificateAuthorityARN", properties.Get("ARN"))
	assert.Equal(t, "certificateAuthorityARN", r.String())
}
