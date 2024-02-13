package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CognitoUserPoolResource = "CognitoUserPool"

func init() {
	registry.Register(&registry.Registration{
		Name:   CognitoUserPoolResource,
		Scope:  nuke.Account,
		Lister: &CognitoUserPoolLister{},
		DependsOn: []string{
			CognitoIdentityPoolResource,
			CognitoUserPoolClientResource,
			CognitoUserPoolDomainResource,
		},
	})
}

type CognitoUserPoolLister struct{}

func (l *CognitoUserPoolLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cognitoidentityprovider.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListUserPools(params)
		if err != nil {
			return nil, err
		}

		for _, pool := range output.UserPools {
			resources = append(resources, &CognitoUserPool{
				svc:  svc,
				name: pool.Name,
				id:   pool.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CognitoUserPool struct {
	svc  *cognitoidentityprovider.CognitoIdentityProvider
	name *string
	id   *string
}

func (f *CognitoUserPool) Remove(_ context.Context) error {
	_, err := f.svc.DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: f.id,
	})

	return err
}

func (f *CognitoUserPool) String() string {
	return *f.name
}
