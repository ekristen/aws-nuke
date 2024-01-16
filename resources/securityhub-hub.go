package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/securityhub"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SecurityHubResource = "SecurityHub"

func init() {
	resource.Register(resource.Registration{
		Name:   SecurityHubResource,
		Scope:  nuke.Account,
		Lister: &SecurityHubLister{},
	})
}

type SecurityHubLister struct{}

func (l *SecurityHubLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := securityhub.New(opts.Session)

	resources := make([]resource.Resource, 0)

	resp, err := svc.DescribeHub(nil)

	if err != nil {
		if awsutil.IsAWSError(err, securityhub.ErrCodeInvalidAccessException) {
			// Security SecurityHub is not enabled for this region
			return resources, nil
		}
		return nil, err
	}

	resources = append(resources, &SecurityHub{
		svc: svc,
		id:  resp.HubArn,
	})
	return resources, nil
}

type SecurityHub struct {
	svc *securityhub.SecurityHub
	id  *string
}

func (hub *SecurityHub) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Arn", hub.id)
	return properties
}

func (hub *SecurityHub) Remove(_ context.Context) error {
	_, err := hub.svc.DisableSecurityHub(&securityhub.DisableSecurityHubInput{})
	return err
}
