package resources

import (
	"context"

	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SNSEndpointResource = "SNSEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:   SNSEndpointResource,
		Scope:  nuke.Account,
		Lister: &SNSEndpointLister{},
	})
}

type SNSEndpointLister struct{}

func (l *SNSEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sns.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var platformApplications []*sns.PlatformApplication

	platformParams := &sns.ListPlatformApplicationsInput{}

	for {
		resp, err := svc.ListPlatformApplications(platformParams)
		if err != nil {
			var awsErr awserr.Error
			ok := errors.As(err, &awsErr)
			if ok && awsErr.Code() == "InvalidAction" && awsErr.Message() == "Operation (ListPlatformApplications) is not supported in this region" {
				// AWS answers with InvalidAction on regions that do not
				// support ListPlatformApplications.
				break
			}

			return nil, err
		}

		for _, platformApplication := range resp.PlatformApplications {
			platformApplications = append(platformApplications, platformApplication)
		}
		if resp.NextToken == nil {
			break
		}

		platformParams.NextToken = resp.NextToken

	}

	params := &sns.ListEndpointsByPlatformApplicationInput{}

	for _, platformApplication := range platformApplications {

		params.PlatformApplicationArn = platformApplication.PlatformApplicationArn

		resp, err := svc.ListEndpointsByPlatformApplication(params)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range resp.Endpoints {
			resources = append(resources, &SNSEndpoint{
				svc: svc,
				ARN: endpoint.EndpointArn,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type SNSEndpoint struct {
	svc *sns.SNS
	ARN *string
}

func (f *SNSEndpoint) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEndpoint(&sns.DeleteEndpointInput{
		EndpointArn: f.ARN,
	})

	return err
}

func (f *SNSEndpoint) String() string {
	return *f.ARN
}
