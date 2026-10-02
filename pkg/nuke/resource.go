package nuke

import (
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
)

// Account is the resource scope that all resources in AWS Nuke are registered against.
const Account registry.Scope = "account"

// ListerOpts are the options for the Lister functions of each individual resource. It is passed in as an interface{}
// so that each implementing tool can define their own options for the lister. Each resource then asserts the type on
// the interface{} to get the options it needs.
type ListerOpts struct {
	Region    *Region
	Session   *session.Session // SDK v1
	Config    *aws.Config      // SDK v2
	AccountID *string
	Logger    *logrus.Entry
}

// CrossServiceConfig returns a config for building a client for a service other than
// the one the resource itself represents, for calls made in support of the resource's
// removal. For example, a regional resource that has to create an IAM role before AWS
// will let it be deleted.
//
// The config handed to a lister carries middleware that keeps a resource type from
// being processed outside the region that owns it. That middleware is attached to the
// config, so it also rejects supporting calls to other services; this strips it.
//
// Use opts.Config, not this, for the resource's own service.
func (o *ListerOpts) CrossServiceConfig() *aws.Config {
	return awsutil.CrossServiceConfig(o.Config)
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

	cfg, err := o.Region.Config(resourceType)
	if err != nil {
		panic(err)
	}

	o.Config = cfg

	if o.Logger != nil {
		o.Logger = o.Logger.WithField("resource", resourceType)
	} else {
		o.Logger = logrus.WithField("resource", resourceType)
	}

	return o
}
