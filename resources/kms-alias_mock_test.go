package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_kmsiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_KMSAlias_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockKMS.EXPECT().ListAliasesPages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(input *kms.ListAliasesInput, fn func(*kms.ListAliasesOutput, bool) bool) error {
			fn(&kms.ListAliasesOutput{
				Aliases: []*kms.AliasListEntry{
					{AliasName: aws.String("alias/test-alias-1")},
					{AliasName: aws.String("alias/test-alias-2")},
				},
			}, true)
			return nil
		},
	)

	lister := KMSAliasLister{
		mockSvc: mockKMS,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.NoError(err)
	a.Len(resources, 2)
}

func Test_Mock_KMSAlias_List_Error(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	mockKMS.EXPECT().
		ListAliasesPages(gomock.Any(), gomock.Any()).
		Return(awserr.New("BadRequest", "400 Bad Request", nil))

	lister := KMSAliasLister{
		mockSvc: mockKMS,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.Error(err)
	a.Nil(resources)
	a.EqualError(err, "BadRequest: 400 Bad Request")
}

func Test_KMSAlias_Filter(t *testing.T) {
	a := assert.New(t)

	alias := &KMSAlias{
		Name: ptr.String("alias/aws/test-alias"),
	}

	err := alias.Filter()
	a.Error(err)
	a.EqualError(err, "cannot delete AWS alias")

	alias.Name = ptr.String("alias/custom/test-alias")
	err = alias.Filter()
	a.NoError(err)
}

func Test_KMSAlias_Properties(t *testing.T) {
	a := assert.New(t)

	alias := &KMSAlias{
		Name: ptr.String("alias/custom/test-alias"),
	}

	a.Equal("alias/custom/test-alias", alias.String())
	a.Equal("alias/custom/test-alias", alias.Properties().Get("Name"))
}

func Test_Mock_KMSAlias_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)

	// Mock the DeleteAlias method
	mockKMS.EXPECT().DeleteAlias(&kms.DeleteAliasInput{
		AliasName: ptr.String("alias/test-alias-1"),
	}).Return(&kms.DeleteAliasOutput{}, nil)

	alias := &KMSAlias{
		svc:  mockKMS,
		Name: ptr.String("alias/test-alias-1"),
	}

	err := alias.Remove(context.TODO())
	a.NoError(err)
}
