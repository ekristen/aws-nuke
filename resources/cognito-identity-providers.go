package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CognitoIdentityProviderResource = "CognitoIdentityProvider"

func init() {
	resource.Register(resource.Registration{
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
			UserPoolId: userPool.id,
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
					userPoolName: userPool.name,
					userPoolId:   userPool.id,
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
	userPoolId   *string
}

func (p *CognitoIdentityProvider) Remove(_ context.Context) error {
	_, err := p.svc.DeleteIdentityProvider(&cognitoidentityprovider.DeleteIdentityProviderInput{
		UserPoolId:   p.userPoolId,
		ProviderName: p.name,
	})

	return err
}

func (p *CognitoIdentityProvider) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Type", p.providerType)
	properties.Set("UserPoolName", p.userPoolName)
	properties.Set("Name", p.name)
	return properties
}

func (p *CognitoIdentityProvider) String() string {
	return *p.userPoolName + " -> " + *p.name
}
