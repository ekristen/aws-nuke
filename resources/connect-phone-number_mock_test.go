package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectPhoneNumber_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectPhoneNumber{
		PhoneNumberID:   ptr.String("phone-id"),
		PhoneNumber:     ptr.String("+15551234567"),
		PhoneNumberType: "DID",
		CountryCode:     "US",
		ARN:             ptr.String("arn:aws:connect:us-east-1:123456789012:phone-number/phone-id"),
		InstanceID:      ptr.String("instance-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("phone-id", props.Get("PhoneNumberID"))
	a.Equal("+15551234567", props.Get("PhoneNumber"))
	a.Equal("DID", props.Get("PhoneNumberType"))
	a.Equal("US", props.Get("CountryCode"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:phone-number/phone-id", props.Get("ARN"))
	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectPhoneNumber_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectPhoneNumber{
		PhoneNumber: ptr.String("+15551234567"),
	}

	a.Equal("+15551234567", resource.String())
}
