package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"                             //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_cognitoidentityprovideriface"
)

func Test_Mock_CognitoUserPool_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_cognitoidentityprovideriface.NewMockCognitoIdentityProviderAPI(ctrl)

	lister := &CognitoUserPoolLister{
		cognitoService: mockSvc,
	}

	mockSvc.EXPECT().ListUserPools(&cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(50),
	}).Return(&cognitoidentityprovider.ListUserPoolsOutput{
		UserPools: []*cognitoidentityprovider.UserPoolDescriptionType{
			{
				Id:   aws.String("test-pool-id"),
				Name: aws.String("test-pool"),
			},
		},
	}, nil)

	mockSvc.EXPECT().ListTagsForResource(&cognitoidentityprovider.ListTagsForResourceInput{
		ResourceArn: aws.String("arn:aws:cognito-idp:us-east-2:012345678901:userpool/test-pool-id"),
	}).Return(&cognitoidentityprovider.ListTagsForResourceOutput{
		Tags: map[string]*string{
			"test-key": aws.String("test-value"),
		},
	}, nil)

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.NoError(err)
	a.Len(resources, 1)
}

func Test_Mock_CognitoUserPool_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_cognitoidentityprovideriface.NewMockCognitoIdentityProviderAPI(ctrl)

	mockSvc.EXPECT().DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String("test-pool-id"),
	}).Return(&cognitoidentityprovider.DeleteUserPoolOutput{}, nil)

	s := &settings.Setting{}
	s.Set("DisableDeletionProtection", false)

	pool := &CognitoUserPool{
		svc:      mockSvc,
		settings: s,
		Name:     aws.String("test-pool"),
		ID:       aws.String("test-pool-id"),
		Tags: map[string]*string{
			"test-key": aws.String("test-value"),
		},
	}

	err := pool.Remove(context.TODO())
	a.NoError(err)
}

func Test_Mock_CognitoUserPool_Remove_DeletionProtection(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_cognitoidentityprovideriface.NewMockCognitoIdentityProviderAPI(ctrl)

	mockSvc.EXPECT().DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String("test-pool-id"),
	}).Return(&cognitoidentityprovider.DescribeUserPoolOutput{
		UserPool: &cognitoidentityprovider.UserPoolType{
			UserAttributeUpdateSettings: &cognitoidentityprovider.UserAttributeUpdateSettingsType{
				AttributesRequireVerificationBeforeUpdate: []*string{ptr.String("email")},
			},
			AutoVerifiedAttributes: []*string{ptr.String("email")},
		},
	}, nil)

	mockSvc.EXPECT().UpdateUserPool(&cognitoidentityprovider.UpdateUserPoolInput{
		UserPoolId:         aws.String("test-pool-id"),
		DeletionProtection: aws.String("INACTIVE"),
		UserAttributeUpdateSettings: &cognitoidentityprovider.UserAttributeUpdateSettingsType{
			AttributesRequireVerificationBeforeUpdate: []*string{ptr.String("email")},
		},
		AutoVerifiedAttributes: []*string{ptr.String("email")},
	}).Return(&cognitoidentityprovider.UpdateUserPoolOutput{}, nil)

	mockSvc.EXPECT().DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String("test-pool-id"),
	}).Return(&cognitoidentityprovider.DeleteUserPoolOutput{}, nil)

	s := &settings.Setting{}
	s.Set("DisableDeletionProtection", true)

	pool := &CognitoUserPool{
		svc:      mockSvc,
		settings: s,
		Name:     aws.String("test-pool"),
		ID:       aws.String("test-pool-id"),
		Tags: map[string]*string{
			"test-key": aws.String("test-value"),
		},
	}

	err := pool.Remove(context.TODO())
	a.NoError(err)
}

func Test_Mock_CognitoUserPool_Properties(t *testing.T) {
	a := assert.New(t)

	s := &settings.Setting{}
	s.Set("DisableDeletionProtection", false)

	pool := &CognitoUserPool{
		settings: s,
		Name:     aws.String("test-pool"),
		ID:       aws.String("test-pool-id"),
		Tags: map[string]*string{
			"test-key": aws.String("test-value"),
		},
	}

	a.Equal("test-pool", pool.Properties().Get("Name"))
	a.Equal("test-pool-id", pool.Properties().Get("ID"))
	a.Equal("test-value", pool.Properties().Get("tag:test-key"))
	a.Equal("test-pool", pool.String())
}
