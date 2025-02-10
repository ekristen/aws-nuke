package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	rtypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func Test_CloudFrontDistribution_Properties(t *testing.T) {
	now := time.Now().UTC()
	r := &CloudFrontDistribution{
		ID:               ptr.String("test-id"),
		Status:           ptr.String("test-status"),
		LastModifiedTime: ptr.Time(now),
		Tags: []rtypes.Tag{
			{
				Key:   ptr.String("test-key"),
				Value: ptr.String("test-value"),
			},
		},
	}
	got := r.Properties()
	assert.Equal(t, "test-id", got.Get("ID"))
	assert.Equal(t, now.Format(time.RFC3339), got.Get("LastModifiedTime"))
	assert.Equal(t, "test-value", got.Get("tag:test-key"))
}
