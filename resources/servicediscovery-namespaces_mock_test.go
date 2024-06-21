package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/servicediscovery"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_servicediscoveryiface"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_ServiceDiscoveryNamespace_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_servicediscoveryiface.NewMockServiceDiscoveryAPI(ctrl)

	resource := ServiceDiscoveryNamespaceLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().ListNamespaces(gomock.Any()).Return(&servicediscovery.ListNamespacesOutput{
		Namespaces: []*servicediscovery.NamespaceSummary{
			{
				Id:  ptr.String("id"),
				Arn: ptr.String("arn:aws:servicediscovery:us-west-2:123456789012:namespace/id"),
			},
		},
	}, nil)

	mockSvc.EXPECT().ListTagsForResource(gomock.Eq(&servicediscovery.ListTagsForResourceInput{
		ResourceARN: ptr.String("arn:aws:servicediscovery:us-west-2:123456789012:namespace/id"),
	})).Return(&servicediscovery.ListTagsForResourceOutput{
		Tags: []*servicediscovery.Tag{
			{
				Key:   ptr.String("foo"),
				Value: ptr.String("bar"),
			},
		},
	}, nil)

	resources, err := resource.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_ServiceDiscoveryNamespace_List_NoTags(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_servicediscoveryiface.NewMockServiceDiscoveryAPI(ctrl)

	resource := ServiceDiscoveryNamespaceLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().ListNamespaces(gomock.Any()).Return(&servicediscovery.ListNamespacesOutput{
		Namespaces: []*servicediscovery.NamespaceSummary{
			{
				Id:  ptr.String("id"),
				Arn: ptr.String("arn:aws:servicediscovery:us-west-2:123456789012:namespace/id"),
			},
		},
	}, nil)

	mockSvc.EXPECT().ListTagsForResource(gomock.Eq(&servicediscovery.ListTagsForResourceInput{
		ResourceARN: ptr.String("arn:aws:servicediscovery:us-west-2:123456789012:namespace/id"),
	})).Return(&servicediscovery.ListTagsForResourceOutput{}, nil)

	resources, err := resource.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_ServiceDiscoveryNamespace_List_TagError(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_servicediscoveryiface.NewMockServiceDiscoveryAPI(ctrl)

	resource := ServiceDiscoveryNamespaceLister{
		mockSvc: mockSvc,
	}

	mockSvc.EXPECT().ListNamespaces(gomock.Any()).Return(&servicediscovery.ListNamespacesOutput{
		Namespaces: []*servicediscovery.NamespaceSummary{
			{
				Id:  ptr.String("id"),
				Arn: ptr.String("arn:aws:servicediscovery:us-west-2:123456789012:namespace/id"),
			},
		},
	}, nil)

	mockSvc.EXPECT().ListTagsForResource(gomock.Eq(&servicediscovery.ListTagsForResourceInput{
		ResourceARN: ptr.String("arn:aws:servicediscovery:us-west-2:123456789012:namespace/id"),
	})).Return(&servicediscovery.ListTagsForResourceOutput{}, fmt.Errorf("error"))

	resources, err := resource.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	namespace := resources[0].(*ServiceDiscoveryNamespace)
	a.Equal("id", namespace.String())
	a.Equal("id", namespace.Properties().Get("ID"))
	a.Equal("", namespace.Properties().Get("tag:foo"))
}

func Test_Mock_ServiceDiscoveryNamespace_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_servicediscoveryiface.NewMockServiceDiscoveryAPI(ctrl)

	resource := ServiceDiscoveryNamespace{
		svc: mockSvc,
		ID:  ptr.String("id"),
	}

	mockSvc.EXPECT().DeleteNamespace(gomock.Eq(&servicediscovery.DeleteNamespaceInput{
		Id: resource.ID,
	})).Return(&servicediscovery.DeleteNamespaceOutput{}, nil)

	err := resource.Remove(context.TODO())
	a.Nil(err)
}
