//go:build integration

package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/awserr" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/session" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck
)

func Test_IAMUser_Remove(t *testing.T) {
	sess := session.Must(session.NewSession())
	svc := iam.New(sess)

	createInput := &iam.CreateUserInput{
		UserName: aws.String("test-user"),
		Tags: []*iam.Tag{
			{
				Key:   aws.String("test-key"),
				Value: aws.String("test-value"),
			},
		},
	}
	out, err := svc.CreateUser(createInput)

	assert.NoError(t, err)
	assert.Equal(t, "test-user", *out.User.UserName)

	iamUser := IAMUser{
		svc:  svc,
		Name: aws.String("test-user"),
		Tags: createInput.Tags,
	}

	removeError := iamUser.Remove(context.TODO())
	assert.NoError(t, removeError)

	_, err = svc.GetUser(&iam.GetUserInput{
		UserName: aws.String("test-user"),
	})
	var awsError awserr.Error
	if errors.As(err, &awsError) {
		assert.Equal(t, "NoSuchEntity", awsError.Code())
	}
}
