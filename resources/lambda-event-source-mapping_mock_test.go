package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_LambdaEventSourceMapping_List(t *testing.T) {
	mockSvc := new(mockLambdaEventSourceMappingClient)
	mockSvc.On("ListEventSourceMappings", mock.Anything, mock.Anything).Return(&lambda.ListEventSourceMappingsOutput{
		EventSourceMappings: []lambdatypes.EventSourceMappingConfiguration{
			{
				UUID:                  ptr.String("test-uuid"),
				EventSourceMappingArn: ptr.String("arn:aws:lambda:us-east-1:123456789012:event-source-mapping:test-uuid"),
				EventSourceArn:        ptr.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
				FunctionArn:           ptr.String("arn:aws:lambda:us-east-1:123456789012:function:test-func"),
				State:                 ptr.String("Enabled"),
			},
		},
	}, nil)
	mockSvc.On("ListTags", mock.Anything, mock.Anything).Return(&lambda.ListTagsOutput{
		Tags: map[string]string{"test-key": "test-value"},
	}, nil)

	lister := &LambdaEventSourceMappingLister{
		mockSvc: mockSvc,
	}
	opts := &nuke.ListerOpts{Config: &aws.Config{
		Region: "us-east-1",
	}}

	resources, err := lister.List(context.TODO(), opts)
	assert.NoError(t, err)
	assert.Len(t, resources, 1)

	res := resources[0].(*LambdaEventSourceMapping)
	assert.Equal(t, "test-uuid", *res.UUID)
	assert.Equal(t, "test-value", res.Tags["test-key"])

	mockSvc.AssertExpectations(t)
}

func Test_Mock_LambdaEventSourceMapping_Remove(t *testing.T) {
	mockSvc := new(mockLambdaEventSourceMappingClient)
	mockSvc.On("DeleteEventSourceMapping", mock.Anything, mock.Anything).Return(&lambda.DeleteEventSourceMappingOutput{}, nil)

	r := &LambdaEventSourceMapping{
		svc:  mockSvc,
		UUID: ptr.String("test-uuid"),
	}

	err := r.Remove(context.TODO())
	assert.NoError(t, err)

	mockSvc.AssertExpectations(t)
}

func Test_Mock_LambdaEventSourceMapping_Properties(t *testing.T) {
	r := &LambdaEventSourceMapping{
		UUID:                  ptr.String("test-uuid"),
		EventSourceMappingArn: ptr.String("arn:aws:lambda:us-east-1:123456789012:event-source-mapping:test-uuid"),
		EventSourceArn:        ptr.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
		FunctionArn:           ptr.String("arn:aws:lambda:us-east-1:123456789012:function:test-func"),
		State:                 ptr.String("Enabled"),
		Tags:                  map[string]string{"env": "test"},
	}

	props := r.Properties()

	assert.Equal(t, "test-uuid", props.Get("UUID"))
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:event-source-mapping:test-uuid", props.Get("EventSourceMappingArn"))
	assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:test-queue", props.Get("EventSourceArn"))
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:test-func", props.Get("FunctionArn"))
	assert.Equal(t, "Enabled", props.Get("State"))
	assert.Equal(t, "test", props.Get("tag:env"))
}
