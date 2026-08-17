package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/smithy-go"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_rdsv2"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	testGlobalClusterID  = "example"
	testGlobalClusterARN = "arn:aws:rds::012345678901:global-cluster:example"
	testWriterARN        = "arn:aws:rds:us-east-2:012345678901:cluster:example-primary"
	testReaderARN        = "arn:aws:rds:eu-central-1:012345678901:cluster:example-secondary"
	testStatusAvailable  = "available"
)

func testGlobalClusterMember(clusterARN string, isWriter bool) rdstypes.GlobalClusterMember {
	return rdstypes.GlobalClusterMember{
		DBClusterArn: ptr.String(clusterARN),
		IsWriter:     ptr.Bool(isWriter),
	}
}

func testDescribeOutput(members ...rdstypes.GlobalClusterMember) *rds.DescribeGlobalClustersOutput {
	return &rds.DescribeGlobalClustersOutput{
		GlobalClusters: []rdstypes.GlobalCluster{
			{
				GlobalClusterIdentifier: ptr.String(testGlobalClusterID),
				GlobalClusterArn:        ptr.String(testGlobalClusterARN),
				Engine:                  ptr.String("aurora-postgresql"),
				EngineVersion:           ptr.String("17.5"),
				Status:                  ptr.String(testStatusAvailable),
				DeletionProtection:      ptr.Bool(false),
				GlobalClusterMembers:    members,
			},
		},
	}
}

func Test_Mock_RDSGlobalCluster_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	// The second global cluster has its writer member in another region and must not be listed there, otherwise every
	// region would try to remove the same global cluster.
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&rds.DescribeGlobalClustersOutput{
			GlobalClusters: []rdstypes.GlobalCluster{
				testDescribeOutput(
					testGlobalClusterMember(testWriterARN, true),
					testGlobalClusterMember(testReaderARN, false),
				).GlobalClusters[0],
				{
					GlobalClusterIdentifier: ptr.String("other"),
					GlobalClusterArn:        ptr.String("arn:aws:rds::012345678901:global-cluster:other"),
					GlobalClusterMembers: []rdstypes.GlobalClusterMember{
						testGlobalClusterMember(testReaderARN, true),
					},
				},
			},
		}, nil)

	mockSvc.EXPECT().ListTagsForResource(gomock.Any(), gomock.Eq(&rds.ListTagsForResourceInput{
		ResourceName: ptr.String(testGlobalClusterARN),
	})).Return(&rds.ListTagsForResourceOutput{
		TagList: []rdstypes.Tag{
			{Key: ptr.String("environment"), Value: ptr.String("sandbox")},
		},
	}, nil)

	lister := &RDSGlobalClusterLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	globalCluster := resources[0].(*RDSGlobalCluster)
	a.Equal(testGlobalClusterID, *globalCluster.Identifier)
	a.Equal("aurora-postgresql", *globalCluster.Engine)
	a.Equal(2, *globalCluster.Members)
	a.Equal("sandbox", globalCluster.Tags["environment"])
}

func Test_Mock_RDSGlobalCluster_List_TagsError(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		testDescribeOutput(testGlobalClusterMember(testWriterARN, true)), nil)
	mockSvc.EXPECT().ListTagsForResource(gomock.Any(), gomock.Any()).Return(
		nil, &rdstypes.GlobalClusterNotFoundFault{})

	lister := &RDSGlobalClusterLister{svc: mockSvc}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region:    &nuke.Region{Name: "us-east-2"},
		AccountID: ptr.String("012345678901"),
		Logger:    logrus.NewEntry(logrus.StandardLogger()),
	})
	a.Nil(err)
	a.Len(resources, 1)
	a.Nil(resources[0].(*RDSGlobalCluster).Tags)
}

func Test_Mock_RDSGlobalCluster_Remove_DetachesReadersOnly(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	// AWS rejects the removal of the writer member while another member is attached, so only the reader is detached.
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Eq(&rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: ptr.String(testGlobalClusterID),
	})).Return(testDescribeOutput(
		testGlobalClusterMember(testWriterARN, true),
		testGlobalClusterMember(testReaderARN, false),
	), nil)

	mockSvc.EXPECT().RemoveFromGlobalCluster(gomock.Any(), gomock.Eq(&rds.RemoveFromGlobalClusterInput{
		GlobalClusterIdentifier: ptr.String(testGlobalClusterID),
		DbClusterIdentifier:     ptr.String(testReaderARN),
	})).Return(&rds.RemoveFromGlobalClusterOutput{}, nil)

	globalCluster := &RDSGlobalCluster{
		svc:        mockSvc,
		settings:   &libsettings.Setting{},
		Identifier: ptr.String(testGlobalClusterID),
	}

	a.Nil(globalCluster.Remove(context.TODO()))
}

func Test_Mock_RDSGlobalCluster_HandleWait(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	globalCluster := &RDSGlobalCluster{
		svc:        mockSvc,
		settings:   &libsettings.Setting{},
		Identifier: ptr.String(testGlobalClusterID),
	}

	// First wait: the reader is gone, so the writer is detached now.
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(
		testDescribeOutput(testGlobalClusterMember(testWriterARN, true)), nil)
	mockSvc.EXPECT().RemoveFromGlobalCluster(gomock.Any(), gomock.Eq(&rds.RemoveFromGlobalClusterInput{
		GlobalClusterIdentifier: ptr.String(testGlobalClusterID),
		DbClusterIdentifier:     ptr.String(testWriterARN),
	})).Return(&rds.RemoveFromGlobalClusterOutput{}, nil)

	var waitErr liberrors.ErrWaitResource
	a.ErrorAs(globalCluster.HandleWait(context.TODO()), &waitErr)

	// Second wait: the removal is still in flight, the global cluster is not empty yet.
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(
		testDescribeOutput(testGlobalClusterMember(testWriterARN, true)), nil)

	a.ErrorAs(globalCluster.HandleWait(context.TODO()), &waitErr)

	// Third wait: no members left, the global cluster is deleted.
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(testDescribeOutput(), nil)
	mockSvc.EXPECT().DeleteGlobalCluster(gomock.Any(), gomock.Eq(&rds.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: ptr.String(testGlobalClusterID),
	})).Return(&rds.DeleteGlobalClusterOutput{}, nil)

	a.Nil(globalCluster.HandleWait(context.TODO()))
}

func Test_Mock_RDSGlobalCluster_HandleWait_DeleteNotEmpty(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(testDescribeOutput(), nil)
	mockSvc.EXPECT().DeleteGlobalCluster(gomock.Any(), gomock.Any()).Return(
		nil, &rdstypes.InvalidGlobalClusterStateFault{})

	globalCluster := &RDSGlobalCluster{
		svc:        mockSvc,
		settings:   &libsettings.Setting{},
		Identifier: ptr.String(testGlobalClusterID),
	}

	var waitErr liberrors.ErrWaitResource
	a.ErrorAs(globalCluster.HandleWait(context.TODO()), &waitErr)
}

func Test_Mock_RDSGlobalCluster_Remove_AlreadyGone(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Times(2).Return(
		nil, &rdstypes.GlobalClusterNotFoundFault{})

	globalCluster := &RDSGlobalCluster{
		svc:        mockSvc,
		settings:   &libsettings.Setting{},
		Identifier: ptr.String(testGlobalClusterID),
	}

	a.Nil(globalCluster.Remove(context.TODO()))
	a.Nil(globalCluster.HandleWait(context.TODO()))
}

func Test_Mock_RDSGlobalCluster_Remove_MemberAlreadyDetached(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(
		testDescribeOutput(testGlobalClusterMember(testReaderARN, false)), nil)
	mockSvc.EXPECT().RemoveFromGlobalCluster(gomock.Any(), gomock.Any()).Return(nil, &smithy.GenericAPIError{
		Code:    "InvalidParameterValue",
		Message: "DBCluster " + testReaderARN + " is not found in global cluster " + testGlobalClusterID,
	})

	globalCluster := &RDSGlobalCluster{
		svc:        mockSvc,
		settings:   &libsettings.Setting{},
		Identifier: ptr.String(testGlobalClusterID),
	}

	a.Nil(globalCluster.Remove(context.TODO()))
}

func Test_Mock_RDSGlobalCluster_Remove_DisableDeletionProtection(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	mockSvc.EXPECT().ModifyGlobalCluster(gomock.Any(), gomock.Eq(&rds.ModifyGlobalClusterInput{
		GlobalClusterIdentifier: ptr.String(testGlobalClusterID),
		DeletionProtection:      ptr.Bool(false),
	})).Return(&rds.ModifyGlobalClusterOutput{}, nil)
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(testDescribeOutput(), nil)
	mockSvc.EXPECT().DeleteGlobalCluster(gomock.Any(), gomock.Any()).Return(&rds.DeleteGlobalClusterOutput{}, nil)

	globalCluster := &RDSGlobalCluster{
		svc:                mockSvc,
		Identifier:         ptr.String(testGlobalClusterID),
		DeletionProtection: ptr.Bool(true),
		settings: &libsettings.Setting{
			"DisableDeletionProtection": true,
		},
	}

	a.Nil(globalCluster.Remove(context.TODO()))
}

func Test_Mock_RDSGlobalCluster_Remove_KeepsDeletionProtection(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_rdsv2.NewMockRDSAPI(ctrl)

	// Without the setting the protection stays, and the delete call fails on the AWS side.
	mockSvc.EXPECT().DescribeGlobalClusters(gomock.Any(), gomock.Any()).Return(testDescribeOutput(), nil)
	mockSvc.EXPECT().DeleteGlobalCluster(gomock.Any(), gomock.Any()).Return(
		nil, &rdstypes.InvalidGlobalClusterStateFault{})

	globalCluster := &RDSGlobalCluster{
		svc:                mockSvc,
		Identifier:         ptr.String(testGlobalClusterID),
		DeletionProtection: ptr.Bool(true),
		settings:           &libsettings.Setting{},
	}

	a.Nil(globalCluster.Remove(context.TODO()))
}

func Test_Mock_RDSGlobalCluster_Filter(t *testing.T) {
	a := assert.New(t)

	a.NoError((&RDSGlobalCluster{Status: ptr.String(testStatusAvailable)}).Filter())
	a.Error((&RDSGlobalCluster{Status: ptr.String(rdsGlobalClusterStatusDeleting)}).Filter())
}

func Test_Mock_RDSGlobalCluster_Properties(t *testing.T) {
	a := assert.New(t)

	globalCluster := &RDSGlobalCluster{
		Identifier:         ptr.String(testGlobalClusterID),
		Engine:             ptr.String("aurora-postgresql"),
		EngineVersion:      ptr.String("17.5"),
		Status:             ptr.String(testStatusAvailable),
		DeletionProtection: ptr.Bool(false),
		Members:            ptr.Int(2),
		Tags: map[string]string{
			"environment": "sandbox",
		},
	}

	properties := globalCluster.Properties()

	a.Equal(testGlobalClusterID, properties.Get("Identifier"))
	a.Equal("aurora-postgresql", properties.Get("Engine"))
	a.Equal("17.5", properties.Get("EngineVersion"))
	a.Equal(testStatusAvailable, properties.Get("Status"))
	a.Equal("false", properties.Get("DeletionProtection"))
	a.Equal("2", properties.Get("Members"))
	a.Equal("sandbox", properties.Get("tag:environment"))
	a.Equal(testGlobalClusterID, globalCluster.String())
}

func Test_Mock_RDSGlobalCluster_BelongsToRegion(t *testing.T) {
	a := assert.New(t)

	lister := &RDSGlobalClusterLister{}

	writerHere := &rdstypes.GlobalCluster{GlobalClusterMembers: []rdstypes.GlobalClusterMember{
		testGlobalClusterMember(testWriterARN, true),
		testGlobalClusterMember(testReaderARN, false),
	}}
	a.True(lister.belongsToRegion(writerHere, "us-east-2"))
	a.False(lister.belongsToRegion(writerHere, "eu-central-1"))

	// Without a writer member the region cannot be determined, so the global cluster is listed everywhere.
	noWriter := &rdstypes.GlobalCluster{GlobalClusterMembers: []rdstypes.GlobalClusterMember{
		testGlobalClusterMember(testReaderARN, false),
	}}
	a.True(lister.belongsToRegion(noWriter, "us-east-2"))
	a.True(lister.belongsToRegion(&rdstypes.GlobalCluster{}, "us-east-2"))

	unparsable := &rdstypes.GlobalCluster{GlobalClusterMembers: []rdstypes.GlobalClusterMember{
		{DBClusterArn: ptr.String("not-an-arn"), IsWriter: ptr.Bool(true)},
	}}
	a.True(lister.belongsToRegion(unparsable, "us-east-2"))
}
