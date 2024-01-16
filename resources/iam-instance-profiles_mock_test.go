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

	err := iamInstanceProfile.Remove(context.TODO())
	a.Nil(err)
}
