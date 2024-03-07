package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/mocks/mock_sagemakeriface"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

// TestSageMakerDomain_List is a unit test function to test the list of SageMakerDomain via mocked interface
func TestSageMakerDomain_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSageMaker := mock_sagemakeriface.NewMockSageMakerAPI(ctrl)

	sagemakerDomainLister := SageMakerDomainLister{
		svc: mockSageMaker,
	}

	sagemakerDomain := SageMakerDomain{
		svc:      mockSageMaker,
		domainID: ptr.String("test"),
		tags: []*sagemaker.Tag{
			{
				Key:   ptr.String("testKey"),
				Value: ptr.String("testValue"),
			},
		},
	}

	mockSageMaker.EXPECT().ListDomains(gomock.Eq(&sagemaker.ListDomainsInput{
		MaxResults: ptr.Int64(30),
	})).Return(&sagemaker.ListDomainsOutput{
		Domains: []*sagemaker.DomainDetails{
			{
				DomainId:  ptr.String("test"),
				DomainArn: ptr.String("testArn"),
			},
		},
	}, nil)

	mockSageMaker.EXPECT().ListTags(gomock.Eq(&sagemaker.ListTagsInput{
		ResourceArn: ptr.String("testArn"),
	})).Return(&sagemaker.ListTagsOutput{
		Tags: []*sagemaker.Tag{
			{
				Key:   ptr.String("testKey"),
				Value: ptr.String("testValue"),
			},
		},
	}, nil)

	resources, err := sagemakerDomainLister.List(context.TODO(), &nuke.ListerOpts{})
	a.NoError(err)
	a.Len(resources, 1)
	a.Equal([]resource.Resource{&sagemakerDomain}, resources)
}

// TestSageMakerDomain_Remove is a unit test function to test the remove of a SageMakerDomain via mocked interface
func TestSageMakerDomain_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSageMaker := mock_sagemakeriface.NewMockSageMakerAPI(ctrl)

	testTime := time.Now().UTC()

	sagemakerDomain := SageMakerDomain{
		svc:          mockSageMaker,
		domainID:     ptr.String("test"),
		creationTime: ptr.Time(testTime),
		tags: []*sagemaker.Tag{
			{
				Key:   ptr.String("testKey"),
				Value: ptr.String("testValue"),
			},
		},
	}

	a.Equal("test", sagemakerDomain.String())
	a.Equal(testTime.Format(time.RFC3339), sagemakerDomain.Properties().Get("CreationTime"))
	a.Equal("testValue", sagemakerDomain.Properties().Get("tag:testKey"))

	mockSageMaker.EXPECT().DeleteDomain(gomock.Eq(&sagemaker.DeleteDomainInput{
		DomainId: sagemakerDomain.domainID,
		RetentionPolicy: &sagemaker.RetentionPolicy{
			HomeEfsFileSystem: ptr.String(sagemaker.RetentionTypeDelete),
		},
	}))

	a.NoError(sagemakerDomain.Remove(context.TODO()))
}
