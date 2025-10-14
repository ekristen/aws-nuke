package resources

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"                             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CognitoUserPoolClientResource = "CognitoUserPoolClient"

func init() {
	registry.Register(&registry.Registration{
		Name:     CognitoUserPoolClientResource,
		Scope:    nuke.Account,
		Resource: &CognitoUserPoolClient{},
		Lister:   &CognitoUserPoolClientLister{},
	})
}

type CognitoUserPoolClientLister struct{}

func (l *CognitoUserPoolClientLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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

		listParams := &cognitoidentityprovider.ListUserPoolClientsInput{
			UserPoolId: userPool.ID,
			MaxResults: aws.Int64(50),
		}

		for {
			output, err := svc.ListUserPoolClients(listParams)
			if err != nil {
				return nil, err
			}

			for _, client := range output.UserPoolClients {
				resources = append(resources, &CognitoUserPoolClient{
					svc:          svc,
					id:           client.ClientId,
					name:         client.ClientName,
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

type CognitoUserPoolClient struct {
	svc          *cognitoidentityprovider.CognitoIdentityProvider
	name         *string
	id           *string
	userPoolName *string
	userPoolID   *string
}

func (r *CognitoUserPoolClient) Remove(_ context.Context) error {
	_, err := r.svc.DeleteUserPoolClient(&cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   r.id,
		UserPoolId: r.userPoolID,
	})

	return err
}

func (r *CognitoUserPoolClient) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", r.id)
	properties.Set("Name", r.name)
	properties.Set("UserPoolName", r.userPoolName)
	return properties
}

func (r *CognitoUserPoolClient) String() string {
	return fmt.Sprintf("%s -> %s", *r.userPoolName, *r.name)
}
