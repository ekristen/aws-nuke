package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SSMActivationResource = "SSMActivation"

func init() {
	resource.Register(&resource.Registration{
		Name:   SSMActivationResource,
		Scope:  nuke.Account,
		Lister: &SSMActivationLister{},
	})
}

type SSMActivationLister struct{}

func (l *SSMActivationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ssm.DescribeActivationsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeActivations(params)
		if err != nil {
			return nil, err
		}

		for _, activation := range output.ActivationList {
			resources = append(resources, &SSMActivation{
				svc: svc,
				ID:  activation.ActivationId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMActivation struct {
	svc *ssm.SSM
	ID  *string
}

func (f *SSMActivation) Remove(_ context.Context) error {
	_, err := f.svc.DeleteActivation(&ssm.DeleteActivationInput{
		ActivationId: f.ID,
	})

	return err
}

func (f *SSMActivation) String() string {
	return *f.ID
}
