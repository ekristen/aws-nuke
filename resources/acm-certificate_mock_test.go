package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/acm" //nolint:staticcheck
)

func Test_Mock_ACMCertificate_Properties(t *testing.T) {
	now := time.Now().UTC()
	acmCertificate := &ACMCertificate{
		ARN:        ptr.String("certificateARN"),
		DomainName: ptr.String("domainName"),
		Status:     ptr.String("status"),
		CreatedAt:  ptr.Time(now),
		Tags: []*acm.Tag{
			{
				Key:   ptr.String("key"),
				Value: ptr.String("value"),
			},
		},
	}

	properties := acmCertificate.Properties()

	assert.Equal(t, "domainName", properties.Get("DomainName"))
	assert.Equal(t, "value", properties.Get("tag:key"))
	assert.Equal(t, now.Format(time.RFC3339), properties.Get("CreatedAt"))
	assert.Equal(t, "status", properties.Get("Status"))
}
