package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
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

type CognitoUserPoolLister struct {
	stsService stsiface.STSAPI
}

func (l *CognitoUserPoolLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var stsSvc stsiface.STSAPI
	if l.stsService != nil {
		stsSvc = l.stsService
	} else {
		stsSvc = sts.New(opts.Session)
	}

	svc := cognitoidentityprovider.New(opts.Session)
	resources := make([]resource.Resource, 0)

	identityOutput, err := stsSvc.GetCallerIdentity(nil)
	if err != nil {
		return nil, err
	}
	accountID := identityOutput.Account

	params := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListUserPools(params)
		if err != nil {
			return nil, err
		}

		for _, pool := range output.UserPools {
			tagResp, tagsErr := svc.ListTagsForResource(&cognitoidentityprovider.ListTagsForResourceInput{
				ResourceArn: ptr.String(fmt.Sprintf("arn:aws:cognito-idp:%s:%s:userpool/%s", opts.Region.Name, *accountID, *pool.Id)),
			})

			if tagsErr != nil {
				logrus.WithError(tagsErr).Error("unable to get tags for userpool")
			}

			resources = append(resources, &CognitoUserPool{
				svc:  svc,
				Name: pool.Name,
				ID:   pool.Id,
				Tags: tagResp.Tags,
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
	Name *string
	ID   *string
	Tags map[string]*string
}

func (f *CognitoUserPool) Remove(_ context.Context) error {
	_, err := f.svc.DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: f.ID,
	})

	return err
}

func (f *CognitoUserPool) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *CognitoUserPool) String() string {
	return *f.Name
}
