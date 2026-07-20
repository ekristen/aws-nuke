package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	kinesistypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_kinesisv2"
)

func Test_Mock_KinesisStream_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_kinesisv2.NewMockKinesisStreamAPI(ctrl)

	mockSvc.EXPECT().ListStreams(gomock.Any(), gomock.Any()).Return(&kinesis.ListStreamsOutput{
		HasMoreStreams: ptr.Bool(false),
		StreamNames:    []string{"stream-1", "stream-2"},
	}, nil)

	mockSvc.EXPECT().ListTagsForStream(gomock.Any(), &kinesis.ListTagsForStreamInput{
		StreamName: ptr.String("stream-1"),
	}).Return(&kinesis.ListTagsForStreamOutput{
		Tags: []kinesistypes.Tag{
			{Key: ptr.String("Environment"), Value: ptr.String("test")},
		},
	}, nil)

	mockSvc.EXPECT().ListTagsForStream(gomock.Any(), &kinesis.ListTagsForStreamInput{
		StreamName: ptr.String("stream-2"),
	}).Return(&kinesis.ListTagsForStreamOutput{
		Tags: []kinesistypes.Tag{},
	}, nil)

	lister := &KinesisStreamLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	stream1 := resources[0].(*KinesisStream)
	a.Equal("stream-1", *stream1.Name)
	a.Equal(map[string]string{"Environment": "test"}, stream1.Tags)

	stream2 := resources[1].(*KinesisStream)
	a.Equal("stream-2", *stream2.Name)
	a.Equal(map[string]string{}, stream2.Tags)
}

func Test_Mock_KinesisStream_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_kinesisv2.NewMockKinesisStreamAPI(ctrl)

	stream := &KinesisStream{
		svc:  mockSvc,
		Name: ptr.String("test-stream"),
	}

	mockSvc.EXPECT().DeleteStream(gomock.Any(), &kinesis.DeleteStreamInput{
		StreamName: stream.Name,
	}).Return(&kinesis.DeleteStreamOutput{}, nil)

	err := stream.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_KinesisStream_Properties(t *testing.T) {
	a := assert.New(t)

	stream := &KinesisStream{
		Name: ptr.String("test-stream"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := stream.Properties()

	a.Equal("test-stream", props.Get("Name"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_KinesisStream_String(t *testing.T) {
	a := assert.New(t)

	stream := &KinesisStream{
		Name: ptr.String("test-stream"),
	}

	a.Equal("test-stream", stream.String())
}
