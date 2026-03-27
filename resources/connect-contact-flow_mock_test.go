package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	connecttypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
)

func Test_ConnectContactFlow_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlow{
		InstanceID:       ptr.String("instance-id"),
		ContactFlowID:    ptr.String("flow-id"),
		Name:             ptr.String("test-flow"),
		ContactFlowType:  "CONTACT_FLOW",
		ContactFlowState: "ACTIVE",
		ARN:              ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/contact-flow/flow-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("flow-id", props.Get("ContactFlowID"))
	a.Equal("test-flow", props.Get("Name"))
	a.Equal("CONTACT_FLOW", props.Get("ContactFlowType"))
	a.Equal("ACTIVE", props.Get("ContactFlowState"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/contact-flow/flow-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectContactFlow_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlow{
		Name: ptr.String("test-flow"),
	}

	a.Equal("test-flow", resource.String())
}

func Test_ConnectContactFlow_Filter_SystemManaged(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlow{
		Name:            ptr.String("default-queue-flow"),
		ContactFlowType: string(connecttypes.ContactFlowTypeCustomerQueue),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "cannot delete system-managed contact flow type")
}

func Test_ConnectContactFlow_Filter_ContactFlow(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlow{
		Name:            ptr.String("my-flow"),
		ContactFlowType: string(connecttypes.ContactFlowTypeContactFlow),
	}

	err := resource.Filter()
	a.Nil(err)
}

func Test_ConnectContactFlow_Filter_Campaign(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlow{
		Name:            ptr.String("my-campaign"),
		ContactFlowType: string(connecttypes.ContactFlowTypeCampaign),
	}

	err := resource.Filter()
	a.Nil(err)
}
