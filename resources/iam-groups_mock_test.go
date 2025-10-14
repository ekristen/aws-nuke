package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
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

	err := iamGroup.Remove(context.TODO())
	a.Nil(err)
}
