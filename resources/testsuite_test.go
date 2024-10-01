package resources

import (
	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

var testListerOpts = &nuke.ListerOpts{
	Region: &nuke.Region{
		Name: "us-east-2",
	},
	Session:   session.Must(session.NewSession()),
	AccountID: ptr.String("012345678901"),
}
