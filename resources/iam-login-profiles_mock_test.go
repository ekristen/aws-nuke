package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
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

	err := iamLoginProfile.Remove()
	a.Nil(err)
}
