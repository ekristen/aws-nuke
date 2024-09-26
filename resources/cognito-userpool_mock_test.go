package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_cognitoidentityprovideriface"
	"github.com/ekristen/aws-nuke/v3/mocks/mock_stsiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_CognitoUserPool_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_cognitoidentityprovideriface.NewMockCognitoIdentityProviderAPI(ctrl)
	mockStsSvc := mock_stsiface.NewMockSTSAPI(ctrl)

	lister := &CognitoUserPoolLister{
		stsService:     mockStsSvc,
		cognitoService: mockSvc,
	}

	mockStsSvc.EXPECT().GetCallerIdentity(gomock.Any()).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}, nil)

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
		ResourceArn: aws.String("arn:aws:cognito-idp:us-east-2:123456789012:userpool/test-pool-id"),
	}).Return(&cognitoidentityprovider.ListTagsForResourceOutput{
		Tags: map[string]*string{
			"test-key": aws.String("test-value"),
		},
	}, nil)

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
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

	mockSvc.EXPECT().UpdateUserPool(&cognitoidentityprovider.UpdateUserPoolInput{
		UserPoolId:         aws.String("test-pool-id"),
		DeletionProtection: aws.String("INACTIVE"),
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
