//go:build integration

package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
)

func Test_KMSKey_Remove(t *testing.T) {
	cfg := aws.NewConfig()
	cfg.Region = ptr.String("us-east-1")

	sess := session.Must(session.NewSession(cfg))
	svc := kms.New(sess)

	stsSvc := sts.New(sess)
	ident, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	assert.NoError(t, err)

	createInput := &kms.CreateKeyInput{
		Policy: aws.String(fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [	
				{	
					"Effect": "Allow",	
					"Principal": {	
						"AWS": "arn:aws:iam::%s:root"	
					},	
					"Action": "kms:*",	
					"Resource": "*"	
				}
			]
		}`, *ident.Account)),
	}

	out, err := svc.CreateKey(createInput)
	assert.NoError(t, err)
	assert.NotNil(t, out)

	kmsKey := KMSKey{
		svc: svc,
		id:  *out.KeyMetadata.KeyId,
	}

	removeError := kmsKey.Remove(context.TODO())
	assert.NoError(t, removeError)

	_, err = svc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(kmsKey.id),
	})
	var awsError awserr.Error
	if errors.As(err, &awsError) {
		assert.Equal(t, "NoSuchEntity", awsError.Code())
	}
}
