package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/fis"
	fistypes "github.com/aws/aws-sdk-go-v2/service/fis/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_fis"
)

func Test_Mock_FISTargetAccountConfiguration_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_fis.NewMockFISAPI(ctrl)

	mockSvc.EXPECT().
		ListExperimentTemplates(gomock.Any(), gomock.Any()).
		Return(&fis.ListExperimentTemplatesOutput{
			ExperimentTemplates: []fistypes.ExperimentTemplateSummary{
				{
					Id: ptr.String("EXT123"),
				},
			},
		}, nil)

	mockSvc.EXPECT().
		ListTargetAccountConfigurations(gomock.Any(), gomock.Any()).
		Return(&fis.ListTargetAccountConfigurationsOutput{
			TargetAccountConfigurations: []fistypes.TargetAccountConfigurationSummary{
				{
					AccountId:   ptr.String("111122223333"),
					RoleArn:     ptr.String("arn:aws:iam::111122223333:role/FISRole"),
					Description: ptr.String("target account"),
				},
			},
		}, nil)

	lister := &FISTargetAccountConfigurationLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	config := resources[0].(*FISTargetAccountConfiguration)
	a.Equal("EXT123", *config.ExperimentTemplateID)
	a.Equal("111122223333", *config.AccountID)
	a.Equal("arn:aws:iam::111122223333:role/FISRole", *config.RoleARN)
	a.Equal("target account", *config.Description)
}

func Test_Mock_FISTargetAccountConfiguration_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_fis.NewMockFISAPI(ctrl)

	mockSvc.EXPECT().
		DeleteTargetAccountConfiguration(gomock.Any(), &fis.DeleteTargetAccountConfigurationInput{
			ExperimentTemplateId: ptr.String("EXT123"),
			AccountId:            ptr.String("111122223333"),
		}).
		Return(&fis.DeleteTargetAccountConfigurationOutput{}, nil)

	config := &FISTargetAccountConfiguration{
		svc:                  mockSvc,
		ExperimentTemplateID: ptr.String("EXT123"),
		AccountID:            ptr.String("111122223333"),
	}

	err := config.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_FISTargetAccountConfiguration_Properties(t *testing.T) {
	a := assert.New(t)

	config := &FISTargetAccountConfiguration{
		ExperimentTemplateID: ptr.String("EXT123"),
		AccountID:            ptr.String("111122223333"),
		RoleARN:              ptr.String("arn:aws:iam::111122223333:role/FISRole"),
		Description:          ptr.String("target account"),
	}

	props := config.Properties()

	a.Equal("EXT123", props.Get("ExperimentTemplateID"))
	a.Equal("111122223333", props.Get("AccountID"))
	a.Equal("arn:aws:iam::111122223333:role/FISRole", props.Get("RoleARN"))
	a.Equal("target account", props.Get("Description"))
	a.Equal("111122223333", config.String())
}
