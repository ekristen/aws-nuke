package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/glue" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_glueiface"
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

func Test_Mock_GlueSecurityConfiguration_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_glueiface.NewMockGlueAPI(ctrl)

	resource := GlueSecurityConfigurationLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().GetSecurityConfigurations(gomock.Any()).Return(&glue.GetSecurityConfigurationsOutput{
		SecurityConfigurations: []*glue.SecurityConfiguration{
			{
				Name: ptr.String("foobar"),
			},
		},
	}, nil)

	resources, err := resource.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)
	a.Equal("foobar", *resources[0].(*GlueSecurityConfiguration).name)
}

func Test_Mock_GlueSecurityConfiguration_ListNext(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_glueiface.NewMockGlueAPI(ctrl)

	resource := GlueSecurityConfigurationLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().GetSecurityConfigurations(gomock.Any()).Return(&glue.GetSecurityConfigurationsOutput{
		SecurityConfigurations: []*glue.SecurityConfiguration{
			{
				Name: ptr.String("foobar1"),
			},
		},
		NextToken: ptr.String("once"),
	}, nil)

	mockSvc.EXPECT().GetSecurityConfigurations(&glue.GetSecurityConfigurationsInput{
		NextToken: ptr.String("once"),
	}).Return(&glue.GetSecurityConfigurationsOutput{
		SecurityConfigurations: []*glue.SecurityConfiguration{
			{
				Name: ptr.String("foobar2"),
			},
		},
		NextToken: &[]string{""}[0], // empty string to break the loop or nil
	}, nil)

	resources, err := resource.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)
	a.Equal("foobar1", *resources[0].(*GlueSecurityConfiguration).name)
}
