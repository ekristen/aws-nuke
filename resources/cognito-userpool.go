package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CognitoUserPoolResource = "CognitoUserPool"

func init() {
	registry.Register(&registry.Registration{
		Name:   CognitoUserPoolResource,
		Scope:  nuke.Account,
		Lister: &CognitoUserPoolLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
		DependsOn: []string{
			CognitoIdentityPoolResource,
			CognitoUserPoolClientResource,
			CognitoUserPoolDomainResource,
		},
	})
}

type CognitoUserPoolLister struct {
	stsService     stsiface.STSAPI
	cognitoService cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (l *CognitoUserPoolLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var stsSvc stsiface.STSAPI
	if l.stsService != nil {
		stsSvc = l.stsService
	} else {
		stsSvc = sts.New(opts.Session)
	}

	var svc cognitoidentityprovideriface.CognitoIdentityProviderAPI
	if l.cognitoService != nil {
		svc = l.cognitoService
	} else {
		svc = cognitoidentityprovider.New(opts.Session)
	}

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
	svc      cognitoidentityprovideriface.CognitoIdentityProviderAPI
	settings *settings.Setting
	Name     *string
	ID       *string
	Tags     map[string]*string
}

func (r *CognitoUserPool) Remove(_ context.Context) error {
	if r.settings.GetBool("DisableDeletionProtection") {
		_, err := r.svc.UpdateUserPool(&cognitoidentityprovider.UpdateUserPoolInput{
			UserPoolId:         r.ID,
			DeletionProtection: ptr.String("INACTIVE"),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: r.ID,
	})

	return err
}

func (r *CognitoUserPool) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CognitoUserPool) String() string {
	return *r.Name
}

func (r *CognitoUserPool) Settings(setting *settings.Setting) {
	r.settings = setting
}
