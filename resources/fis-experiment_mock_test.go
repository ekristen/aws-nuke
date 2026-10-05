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

func Test_Mock_FISExperiment_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_fis.NewMockFISAPI(ctrl)
	created := time.Now().Add(-time.Hour)

	mockSvc.EXPECT().
		ListExperiments(gomock.Any(), gomock.Any()).
		Return(&fis.ListExperimentsOutput{
			Experiments: []fistypes.ExperimentSummary{
				{
					Id:                   ptr.String("EXP123"),
					Arn:                  ptr.String("arn:aws:fis:us-east-1:123456789012:experiment/EXP123"),
					ExperimentTemplateId: ptr.String("EXT123"),
					CreationTime:         &created,
					State: &fistypes.ExperimentState{
						Status: fistypes.ExperimentStatusRunning,
					},
					Tags: map[string]string{"env": "test"},
				},
				{
					Id:                   ptr.String("EXP456"),
					Arn:                  ptr.String("arn:aws:fis:us-east-1:123456789012:experiment/EXP456"),
					ExperimentTemplateId: ptr.String("EXT123"),
					CreationTime:         &created,
					State: &fistypes.ExperimentState{
						Status: fistypes.ExperimentStatusCompleted,
					},
				},
			},
		}, nil)

	lister := &FISExperimentLister{
		svc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	first := resources[0].(*FISExperiment)
	a.Equal("EXP123", *first.ID)
	a.Equal("EXT123", *first.ExperimentTemplateID)
	a.Equal(string(fistypes.ExperimentStatusRunning), first.Status)
	a.Equal("test", first.Tags["env"])

	second := resources[1].(*FISExperiment)
	a.Equal("EXP456", *second.ID)
	a.Equal(string(fistypes.ExperimentStatusCompleted), second.Status)
}

func Test_Mock_FISExperiment_Filter(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		Name     string
		Status   string
		Filtered bool
	}{
		{Name: "pending", Status: string(fistypes.ExperimentStatusPending), Filtered: false},
		{Name: "initiating", Status: string(fistypes.ExperimentStatusInitiating), Filtered: false},
		{Name: "running", Status: string(fistypes.ExperimentStatusRunning), Filtered: false},
		{Name: "completed", Status: string(fistypes.ExperimentStatusCompleted), Filtered: true},
		{Name: "stopping", Status: string(fistypes.ExperimentStatusStopping), Filtered: true},
		{Name: "stopped", Status: string(fistypes.ExperimentStatusStopped), Filtered: true},
		{Name: "failed", Status: string(fistypes.ExperimentStatusFailed), Filtered: true},
		{Name: "cancelled", Status: string(fistypes.ExperimentStatusCancelled), Filtered: true},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			experiment := &FISExperiment{
				ID:     ptr.String("EXP123"),
				Status: c.Status,
			}

			err := experiment.Filter()
			if c.Filtered {
				a.NotNil(err)
			} else {
				a.Nil(err)
			}
		})
	}
}

func Test_Mock_FISExperiment_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_fis.NewMockFISAPI(ctrl)

	mockSvc.EXPECT().
		StopExperiment(gomock.Any(), &fis.StopExperimentInput{
			Id: ptr.String("EXP123"),
		}).
		Return(&fis.StopExperimentOutput{}, nil)

	experiment := &FISExperiment{
		svc:    mockSvc,
		ID:     ptr.String("EXP123"),
		Status: string(fistypes.ExperimentStatusRunning),
	}

	err := experiment.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_FISExperiment_Properties(t *testing.T) {
	a := assert.New(t)

	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	experiment := &FISExperiment{
		ID:                   ptr.String("EXP123"),
		ARN:                  ptr.String("arn:aws:fis:us-east-1:123456789012:experiment/EXP123"),
		ExperimentTemplateID: ptr.String("EXT123"),
		Status:               string(fistypes.ExperimentStatusRunning),
		CreationTime:         &created,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := experiment.Properties()

	a.Equal("EXP123", props.Get("ID"))
	a.Equal("arn:aws:fis:us-east-1:123456789012:experiment/EXP123", props.Get("ARN"))
	a.Equal("EXT123", props.Get("ExperimentTemplateID"))
	a.Equal("running", props.Get("Status"))
	a.Equal("test", props.Get("tag:Environment"))
	a.Equal("EXP123", experiment.String())
}
