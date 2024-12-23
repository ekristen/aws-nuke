package resources

import (
	"fmt"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestS3AccessPointProperties(t *testing.T) {
	tests := []struct {
		accountID     string
		name          string
		alias         string
		bucket        string
		networkOrigin string
	}{
		{
			accountID:     "123456789012",
			name:          "test-access-point",
			alias:         "some-alias",
			bucket:        "some-bucket",
			networkOrigin: "some-network-origin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &S3AccessPoint{
				accountID:     ptr.String(tc.accountID),
				ARN:           ptr.String(fmt.Sprintf("arn:aws:s3:::%s:%s", tc.accountID, tc.name)),
				Name:          ptr.String(tc.name),
				Alias:         ptr.String(tc.alias),
				Bucket:        ptr.String(tc.bucket),
				NetworkOrigin: ptr.String(tc.networkOrigin),
			}

			got := obj.Properties()
			assert.Equal(t, tc.name, got.Get("Name"))
			assert.Equal(t, fmt.Sprintf("arn:aws:s3:::%s:%s", tc.accountID, tc.name), got.Get("AccessPointArn"))
			assert.Equal(t, tc.alias, got.Get("Alias"))
			assert.Equal(t, tc.bucket, got.Get("Bucket"))
			assert.Equal(t, tc.networkOrigin, got.Get("NetworkOrigin"))

			assert.Equal(t, fmt.Sprintf("arn:aws:s3:::%s:%s", tc.accountID, tc.name), obj.String())
		})
	}
}
