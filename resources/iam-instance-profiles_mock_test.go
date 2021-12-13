package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMInstanceProfile_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamInstanceProfile := IAMInstanceProfile{
		svc:  mockIAM,
		name: "ip:foobar",
	}

	mockIAM.EXPECT().DeleteInstanceProfile(gomock.Eq(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(iamInstanceProfile.name),
	})).Return(&iam.DeleteInstanceProfileOutput{}, nil)

	err := iamInstanceProfile.Remove()
	a.Nil(err)
}
