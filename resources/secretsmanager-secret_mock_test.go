package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_secretsmanageriface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_SecretsManager_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)

	lister := SecretsManagerSecretLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().ListSecrets(gomock.Any()).Return(&secretsmanager.ListSecretsOutput{
		SecretList: []*secretsmanager.SecretListEntry{
			{
				Name: ptr.String("foo"),
				ARN:  ptr.String("arn:foo"),
				Tags: []*secretsmanager.Tag{
					{
						Key:   ptr.String("foo"),
						Value: ptr.String("bar"),
					},
				},
			},
			{
				Name:          ptr.String("bar"),
				ARN:           ptr.String("arn:bar"),
				PrimaryRegion: ptr.String("us-west-2"),
			},
		},
	}, nil)

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession(&aws.Config{})),
	})
	a.Nil(err)
	a.Len(resources, 2)

	resource1 := resources[0].(*SecretsManagerSecret)
	a.Equal("foo", resource1.Properties().Get("Name"))
	a.Equal("arn:foo", resource1.Properties().Get("ARN"))
	a.Equal("bar", resource1.Properties().Get("tag:foo"))

	resource2 := resources[1].(*SecretsManagerSecret)
	a.Equal("bar", resource2.Properties().Get("Name"))
	a.Equal("arn:bar", resource2.Properties().Get("ARN"))
	a.Equal("us-west-2", resource2.Properties().Get("PrimaryRegion"))
}

func Test_Mock_SecretsManager_Secret_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)

	resource := SecretsManagerSecret{
		svc:  mockSvc,
		ARN:  ptr.String("arn:foo"),
		Name: ptr.String("foo"),
	}

	mockSvc.EXPECT().DeleteSecret(gomock.Eq(&secretsmanager.DeleteSecretInput{
		SecretId:                   ptr.String("arn:foo"),
		ForceDeleteWithoutRecovery: ptr.Bool(true),
	})).Return(&secretsmanager.DeleteSecretOutput{}, nil)

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_SecretsManager_Secret_RemoveReplica(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)

	resource := SecretsManagerSecret{
		svc:           mockSvc,
		primarySvc:    mockSvc,
		region:        ptr.String("us-east-1"), // region this replica is in
		ARN:           ptr.String("arn:foo"),
		Name:          ptr.String("foo"),
		PrimaryRegion: ptr.String("us-west-2"),
		Replica:       true,
	}

	mockSvc.EXPECT().RemoveRegionsFromReplication(gomock.Eq(&secretsmanager.RemoveRegionsFromReplicationInput{
		SecretId:             ptr.String("arn:foo"),
		RemoveReplicaRegions: []*string{ptr.String("us-east-1")},
	})).Return(&secretsmanager.RemoveRegionsFromReplicationOutput{}, nil)

	err := resource.Remove(context.TODO())
	a.Nil(err)
}
