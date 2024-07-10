//go:build integration

package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
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
		name: "test-user",
		tags: createInput.Tags,
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
