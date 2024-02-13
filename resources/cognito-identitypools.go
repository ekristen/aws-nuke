package resources

import (
	"context"

	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cognitoidentity"
)

type CognitoIdentityPool struct {
	svc  *cognitoidentity.CognitoIdentity
	name *string
	id   *string
}

const CognitoIdentityPoolResource = "CognitoIdentityPool"

func init() {
	registry.Register(&registry.Registration{
		Name:   CognitoIdentityPoolResource,
		Scope:  nuke.Account,
		Lister: &CognitoIdentityPoolLister{},
	})
}

type CognitoIdentityPoolLister struct{}

func (l *CognitoIdentityPoolLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cognitoidentity.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListIdentityPools(params)
		if err != nil {
			return nil, err
		}

		for _, pool := range output.IdentityPools {
			resources = append(resources, &CognitoIdentityPool{
				svc:  svc,
				name: pool.IdentityPoolName,
				id:   pool.IdentityPoolId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *CognitoIdentityPool) Remove(_ context.Context) error {

	_, err := f.svc.DeleteIdentityPool(&cognitoidentity.DeleteIdentityPoolInput{
		IdentityPoolId: f.id,
	})

	return err
}

func (f *CognitoIdentityPool) String() string {
	return *f.name
}
