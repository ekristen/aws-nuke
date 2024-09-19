package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CognitoIdentityProviderResource = "CognitoIdentityProvider"

func init() {
	registry.Register(&registry.Registration{
		Name:   CognitoIdentityProviderResource,
		Scope:  nuke.Account,
		Lister: &CognitoIdentityProviderLister{},
	})
}

type CognitoIdentityProviderLister struct{}

func (l *CognitoIdentityProviderLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cognitoidentityprovider.New(opts.Session)

	userPoolsLister := &CognitoUserPoolLister{}
	userPools, poolErr := userPoolsLister.List(ctx, o)
	if poolErr != nil {
		return nil, poolErr
	}

	resources := make([]resource.Resource, 0)

	for _, userPoolResource := range userPools {
		userPool, ok := userPoolResource.(*CognitoUserPool)
		if !ok {
			logrus.Errorf("Unable to case CognitoUserPool")
			continue
		}

		listParams := &cognitoidentityprovider.ListIdentityProvidersInput{
			UserPoolId: userPool.ID,
			MaxResults: aws.Int64(50),
		}

		for {
			output, err := svc.ListIdentityProviders(listParams)
			if err != nil {
				return nil, err
			}

			for _, provider := range output.Providers {
				resources = append(resources, &CognitoIdentityProvider{
					svc:          svc,
					name:         provider.ProviderName,
					providerType: provider.ProviderType,
					userPoolName: userPool.Name,
					userPoolID:   userPool.ID,
				})
			}

			if output.NextToken == nil {
				break
			}

			listParams.NextToken = output.NextToken
		}
	}

	return resources, nil
}

type CognitoIdentityProvider struct {
	svc          *cognitoidentityprovider.CognitoIdentityProvider
	name         *string
	providerType *string
	userPoolName *string
	userPoolID   *string
}

func (r *CognitoIdentityProvider) Remove(_ context.Context) error {
	_, err := r.svc.DeleteIdentityProvider(&cognitoidentityprovider.DeleteIdentityProviderInput{
		UserPoolId:   r.userPoolID,
		ProviderName: r.name,
	})

	return err
}

func (r *CognitoIdentityProvider) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Type", r.providerType)
	properties.Set("UserPoolName", r.userPoolName)
	properties.Set("Name", r.name)
	return properties
}

func (r *CognitoIdentityProvider) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(r.userPoolName), ptr.ToString(r.name))
}
