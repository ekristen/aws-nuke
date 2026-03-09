package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectHoursOfOperation_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectHoursOfOperation{
		InstanceID:         ptr.String("instance-id"),
		HoursOfOperationID: ptr.String("hours-id"),
		Name:               ptr.String("custom-hours"),
		ARN:                ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/operating-hours/hours-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("hours-id", props.Get("HoursOfOperationID"))
	a.Equal("custom-hours", props.Get("Name"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/operating-hours/hours-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectHoursOfOperation_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectHoursOfOperation{
		Name: ptr.String("custom-hours"),
	}

	a.Equal("custom-hours", resource.String())
}

func Test_ConnectHoursOfOperation_Filter_Default(t *testing.T) {
	a := assert.New(t)

	resource := ConnectHoursOfOperation{
		Name: ptr.String("Basic Hours"),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "cannot delete default hours of operation")
}

func Test_ConnectHoursOfOperation_Filter_Custom(t *testing.T) {
	a := assert.New(t)

	resource := ConnectHoursOfOperation{
		Name: ptr.String("Extended Hours"),
	}

	err := resource.Filter()
	a.Nil(err)
}
