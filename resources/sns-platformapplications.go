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

const SNSPlatformApplicationResource = "SNSPlatformApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:   SNSPlatformApplicationResource,
		Scope:  nuke.Account,
		Lister: &SNSPlatformApplicationLister{},
	})
}

type SNSPlatformApplicationLister struct{}

func (l *SNSPlatformApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sns.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sns.ListPlatformApplicationsInput{}

	for {
		resp, err := svc.ListPlatformApplications(params)
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
			resources = append(resources, &SNSPlatformApplication{
				svc: svc,
				ARN: platformApplication.PlatformApplicationArn,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type SNSPlatformApplication struct {
	svc *sns.SNS
	ARN *string
}

func (f *SNSPlatformApplication) Remove(_ context.Context) error {
	_, err := f.svc.DeletePlatformApplication(&sns.DeletePlatformApplicationInput{
		PlatformApplicationArn: f.ARN,
	})

	return err
}

func (f *SNSPlatformApplication) String() string {
	return *f.ARN
}
