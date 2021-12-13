package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/rebuy-de/aws-nuke/v2/mocks/mock_iamiface"
	"github.com/stretchr/testify/assert"
)

func Test_Mock_IAMGroup_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamGroup := IAMGroup{
		svc:  mockIAM,
		name: "foobar",
	}

	mockIAM.EXPECT().DeleteGroup(gomock.Eq(&iam.DeleteGroupInput{
		GroupName: aws.String(iamGroup.name),
	})).Return(&iam.DeleteGroupOutput{}, nil)

	err := iamGroup.Remove()
	a.Nil(err)
}
