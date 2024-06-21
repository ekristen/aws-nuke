package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SESIdentityResource = "SESIdentity"

func init() {
	registry.Register(&registry.Registration{
		Name:   SESIdentityResource,
		Scope:  nuke.Account,
		Lister: &SESIdentityLister{},
	})
}

type SESIdentityLister struct{}

func (l *SESIdentityLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ses.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ses.ListIdentitiesInput{
		MaxItems: aws.Int64(100),
	}

	for {
		output, err := svc.ListIdentities(params)
		if err != nil {
			return nil, err
		}

		for _, identity := range output.Identities {
			resources = append(resources, &SESIdentity{
				svc:      svc,
				identity: identity,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SESIdentity struct {
	svc      *ses.SES
	identity *string
}

func (f *SESIdentity) Remove(_ context.Context) error {
	_, err := f.svc.DeleteIdentity(&ses.DeleteIdentityInput{
		Identity: f.identity,
	})

	return err
}

func (f *SESIdentity) String() string {
	return *f.identity
}
