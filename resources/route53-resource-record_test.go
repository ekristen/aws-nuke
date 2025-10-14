package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/route53" //nolint:staticcheck
)

func TestRoute53ResourceRecordSet_Properties(t *testing.T) {
	cases := []struct {
		name       string
		recordType string
	}{
		{
			name:       "example.com",
			recordType: "NS",
		},
		{
			name:       "example.com",
			recordType: "SOA",
		},
		{
			name:       "subdomain.example.com",
			recordType: "A",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Route53ResourceRecordSet{
				resourceRecordSet: &route53.ResourceRecordSet{
					Name: ptr.String(tc.name),
					Type: ptr.String(tc.recordType),
				},
				Name: ptr.String(tc.name),
				Type: ptr.String(tc.recordType),
			}

			got := r.Properties()
			assert.Equal(t, tc.name, got.Get("Name"))
			assert.Equal(t, tc.recordType, got.Get("Type"))

			assert.Equal(t, tc.name, r.String())
		})
	}
}
