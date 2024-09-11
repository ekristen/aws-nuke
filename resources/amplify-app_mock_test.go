package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

var app = &AmplifyApp{
	Name:  ptr.String("app1"),
	AppID: ptr.String("appId1"),
	Tags: map[string]*string{
		"key1": ptr.String("value1"),
	},
}

func Test_AmplifyApp_Properties(t *testing.T) {
	a := assert.New(t)

	properties := app.Properties()
	a.Equal("app1", properties.Get("Name"))
	a.Equal("appId1", properties.Get("AppID"))
	a.Equal("value1", properties.Get("tag:key1"))
}

func Test_AmplifyApp_Stringer(t *testing.T) {
	a := assert.New(t)

	a.Equal("appId1", app.String())
}
