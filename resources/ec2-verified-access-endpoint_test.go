package resources

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessEndpoint_Properties_MinimalData(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		endpoint: &ec2types.VerifiedAccessEndpoint{
			VerifiedAccessEndpointId: ptr.String("vae-1234567890abcdef0"),
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
			VerifiedAccessGroupId:    ptr.String("vag-1234567890abcdef0"),
			EndpointType:             "load-balancer",
			ApplicationDomain:        ptr.String("example.com"),
			EndpointDomain:           ptr.String("test.example.com"),
			Description:              ptr.String("Test verified access endpoint"),
			CreationTime:             ptr.String(now),
			LastUpdatedTime:          ptr.String(now),
			Status: &ec2types.VerifiedAccessEndpointStatus{
				Code: "active",
			},
			AttachmentType:       "",
			DomainCertificateArn: nil,
			Tags:                 []ec2types.Tag{},
		},
		ID:                    ptr.String("vae-1234567890abcdef0"),
		Description:           ptr.String("Test verified access endpoint"),
		CreationTime:          ptr.String(now),
		LastUpdatedTime:       ptr.String(now),
		VerifiedAccessGroupId: ptr.String("vag-1234567890abcdef0"),
		ApplicationDomain:     ptr.String("example.com"),
		EndpointType:          ptr.String("load-balancer"),
		AttachmentType:        ptr.String(""),
		DomainCertificateArn:  nil,
	}

	properties := endpoint.Properties()

	assert.Equal(t, "vae-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "", properties.Get("AttachmentType"))
	assert.Equal(t, "", properties.Get("DomainCertificateArn"))
}

func Test_EC2VerifiedAccessEndpoint_String(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		endpoint: &ec2types.VerifiedAccessEndpoint{
			VerifiedAccessEndpointId: ptr.String("vae-1234567890abcdef0"),
		},
	}

	assert.Equal(t, "vae-1234567890abcdef0", endpoint.String())
}
