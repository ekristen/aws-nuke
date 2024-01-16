package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/mocks/mock_iamiface"
)

func Test_Mock_IAMLoginProfile_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamLoginProfile := IAMLoginProfile{
		svc:  mockIAM,
		name: "login-profile:foobar",
	}

	mockIAM.EXPECT().DeleteLoginProfile(gomock.Eq(&iam.DeleteLoginProfileInput{
		UserName: aws.String(iamLoginProfile.name),
	})).Return(&iam.DeleteLoginProfileOutput{}, nil)

	err := iamLoginProfile.Remove(context.TODO())
	a.Nil(err)
}
