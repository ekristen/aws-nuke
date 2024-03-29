package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/aws-nuke/mocks/mock_glueiface"
)

func Test_Mock_GlueSecurityConfiguration_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_glueiface.NewMockGlueAPI(ctrl)

	resource := GlueSecurityConfiguration{
		svc:  mockSvc,
		name: ptr.String("foobar"),
	}

	mockSvc.EXPECT().DeleteSecurityConfiguration(gomock.Eq(&glue.DeleteSecurityConfigurationInput{
		Name: resource.name,
	})).Return(&glue.DeleteSecurityConfigurationOutput{}, nil)

	err := resource.Remove(context.TODO())
	a.Nil(err)
}
