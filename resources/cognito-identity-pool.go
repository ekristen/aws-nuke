package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                     //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cognitoidentity" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type CognitoIdentityPool struct {
	svc  *cognitoidentity.CognitoIdentity
	name *string
	id   *string
}

const CognitoIdentityPoolResource = "CognitoIdentityPool"

func init() {
	registry.Register(&registry.Registration{
		Name:     CognitoIdentityPoolResource,
		Scope:    nuke.Account,
		Resource: &CognitoIdentityPool{},
		Lister:   &CognitoIdentityPoolLister{},
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

func (r *CognitoIdentityPool) Remove(_ context.Context) error {
	_, err := r.svc.DeleteIdentityPool(&cognitoidentity.DeleteIdentityPoolInput{
		IdentityPoolId: r.id,
	})

	return err
}

func (r *CognitoIdentityPool) String() string {
	return *r.name
}
