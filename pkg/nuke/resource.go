package nuke

import (
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/ekristen/libnuke/pkg/resource"
)

// Account is the resource scope that all resources in AWS Nuke are registered against.
const Account resource.Scope = "account"

// ListerOpts are the options for the Lister functions of each individual resource. It is passed in as an interface{}
// so that each implementing tool can define their own options for the lister. Each resource then asserts the type on
// the interface{} to get the options it needs.
type ListerOpts struct {
	Region  *Region
	Session *session.Session
}

// MutateOpts is a function that will be called for each resource type to mutate the options for the scanner based on
// whatever criteria you want. However, in this case for the aws-nuke tool, it's mutating the opts to create the proper
// session for the proper region for the resourceType. For example IAM only happens in the global region, not us-east-2.
var MutateOpts = func(opts interface{}, resourceType string) interface{} {
	o := opts.(*ListerOpts)

	session, err := o.Region.Session(resourceType)
	if err != nil {
		panic(err)
	}

	o.Session = session
	return o
}
