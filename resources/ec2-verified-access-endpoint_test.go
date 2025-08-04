package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessEndpoint_Properties_MinimalData(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		endpoint: &ec2.VerifiedAccessEndpoint{
			VerifiedAccessEndpointId: ptr.String("vae-1234567890abcdef0"),
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
			VerifiedAccessGroupId:    ptr.String("vag-1234567890abcdef0"),
			EndpointType:             ptr.String("load-balancer"),
			ApplicationDomain:        ptr.String("example.com"),
			EndpointDomain:           ptr.String("test.example.com"),
			Description:              ptr.String("Test verified access endpoint"),
			CreationTime:             ptr.String(now),
			LastUpdatedTime:          ptr.String(now),
			Status: &ec2.VerifiedAccessEndpointStatus{
				Code: ptr.String("active"),
			},
			AttachmentType:       nil,
			DomainCertificateArn: nil,
			Tags:                 []*ec2.Tag{},
		},
	}

	properties := endpoint.Properties()

	assert.Equal(t, "vae-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "", properties.Get("AttachmentType"))
	assert.Equal(t, "", properties.Get("DomainCertificateArn"))
}

func Test_EC2VerifiedAccessEndpoint_String(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		endpoint: &ec2.VerifiedAccessEndpoint{
			VerifiedAccessEndpointId: ptr.String("vae-1234567890abcdef0"),
		},
	}

	assert.Equal(t, "vae-1234567890abcdef0", endpoint.String())
}
