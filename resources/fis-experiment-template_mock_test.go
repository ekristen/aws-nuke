package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/fis"
	fistypes "github.com/aws/aws-sdk-go-v2/service/fis/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_fis"
)

func Test_Mock_FISExperimentTemplate_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_fis.NewMockFISAPI(ctrl)
	created := time.Now().Add(-24 * time.Hour)
	updated := time.Now().Add(-time.Hour)

	mockSvc.EXPECT().
		ListExperimentTemplates(gomock.Any(), gomock.Any()).
		Return(&fis.ListExperimentTemplatesOutput{
			ExperimentTemplates: []fistypes.ExperimentTemplateSummary{
				{
					Id:             ptr.String("EXT123"),
					Arn:            ptr.String("arn:aws:fis:us-east-1:123456789012:experiment-template/EXT123"),
					Description:    ptr.String("stop instances"),
					CreationTime:   &created,
					LastUpdateTime: &updated,
					Tags:           map[string]string{"team": "sre"},
				},
			},
		}, nil)

	lister := &FISExperimentTemplateLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	template := resources[0].(*FISExperimentTemplate)
	a.Equal("EXT123", *template.ID)
	a.Equal("stop instances", *template.Description)
	a.Equal("sre", template.Tags["team"])
}

func Test_Mock_FISExperimentTemplate_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_fis.NewMockFISAPI(ctrl)

	mockSvc.EXPECT().
		DeleteExperimentTemplate(gomock.Any(), &fis.DeleteExperimentTemplateInput{
			Id: ptr.String("EXT123"),
		}).
		Return(&fis.DeleteExperimentTemplateOutput{}, nil)

	template := &FISExperimentTemplate{
		svc: mockSvc,
		ID:  ptr.String("EXT123"),
	}

	err := template.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_FISExperimentTemplate_Properties(t *testing.T) {
	a := assert.New(t)

	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	template := &FISExperimentTemplate{
		ID:             ptr.String("EXT123"),
		ARN:            ptr.String("arn:aws:fis:us-east-1:123456789012:experiment-template/EXT123"),
		Description:    ptr.String("stop instances"),
		CreationTime:   &created,
		LastUpdateTime: &updated,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := template.Properties()

	a.Equal("EXT123", props.Get("ID"))
	a.Equal("arn:aws:fis:us-east-1:123456789012:experiment-template/EXT123", props.Get("ARN"))
	a.Equal("stop instances", props.Get("Description"))
	a.Equal("test", props.Get("tag:Environment"))
	a.Equal("EXT123", template.String())
}
