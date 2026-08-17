//go:build integration

package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const testRDSGlobalClusterRegion = "us-east-1"

// TestRDSGlobalClusterSuite runs against a global cluster without member clusters. Such a global cluster is created
// within seconds and holds no database, so the test needs no Aurora cluster and produces no database cost.
type TestRDSGlobalClusterSuite struct {
	suite.Suite
	ctx             context.Context
	cfg             aws.Config
	svc             *rds.Client
	globalClusterID *string
}

func (suite *TestRDSGlobalClusterSuite) SetupSuite() {
	suite.ctx = context.TODO()
	suite.globalClusterID = ptr.String(fmt.Sprintf("aws-nuke-test-%d", time.Now().UnixNano()))

	cfg, err := config.LoadDefaultConfig(suite.ctx, config.WithRegion(testRDSGlobalClusterRegion))
	if err != nil {
		suite.T().Fatalf("failed to load config, %v", err)
	}

	suite.cfg = cfg
	suite.svc = rds.NewFromConfig(cfg)

	_, err = suite.svc.CreateGlobalCluster(suite.ctx, &rds.CreateGlobalClusterInput{
		GlobalClusterIdentifier: suite.globalClusterID,
		Engine:                  ptr.String("aurora-postgresql"),
	})
	if err != nil {
		suite.T().Fatalf("failed to create test global cluster, %v", err)
	}

	suite.waitForGlobalCluster()
}

// waitForGlobalCluster waits until the new global cluster shows up in the unfiltered DescribeGlobalClusters response,
// which is what the lister reads. A global cluster is returned by the identifier-filtered call right after its
// creation, but it takes longer to appear in the unfiltered list.
func (suite *TestRDSGlobalClusterSuite) waitForGlobalCluster() {
	for i := 0; i < 60; i++ {
		res, err := suite.svc.DescribeGlobalClusters(suite.ctx, &rds.DescribeGlobalClustersInput{})
		if err != nil {
			suite.T().Fatalf("failed to describe global clusters, %v", err)
		}

		for _, globalCluster := range res.GlobalClusters {
			if ptr.ToString(globalCluster.GlobalClusterIdentifier) == ptr.ToString(suite.globalClusterID) {
				return
			}
		}

		time.Sleep(time.Second)
	}

	suite.T().Fatalf("test global cluster %s did not appear", ptr.ToString(suite.globalClusterID))
}

func (suite *TestRDSGlobalClusterSuite) TearDownSuite() {
	_, _ = suite.svc.DeleteGlobalCluster(suite.ctx, &rds.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: suite.globalClusterID,
	})
}

func (suite *TestRDSGlobalClusterSuite) TestList() {
	a := assert.New(suite.T())

	lister := &RDSGlobalClusterLister{}
	resources, err := lister.List(suite.ctx, &nuke.ListerOpts{
		Config: &suite.cfg,
		Region: &nuke.Region{Name: testRDSGlobalClusterRegion},
		Logger: logrus.NewEntry(logrus.StandardLogger()),
	})

	a.NoError(err)
	a.NotNil(suite.find(resources))
}

func (suite *TestRDSGlobalClusterSuite) TestRemove() {
	a := assert.New(suite.T())

	globalCluster := &RDSGlobalCluster{
		svc:        suite.svc,
		Identifier: suite.globalClusterID,
	}

	a.NoError(globalCluster.Remove(suite.ctx))

	_, err := suite.svc.DescribeGlobalClusters(suite.ctx, &rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: suite.globalClusterID,
	})
	a.True(isGlobalClusterNotFound(err), "global cluster still exists: %v", err)
}

func (suite *TestRDSGlobalClusterSuite) find(resources []resource.Resource) *RDSGlobalCluster {
	for _, r := range resources {
		globalCluster, ok := r.(*RDSGlobalCluster)
		if ok && ptr.ToString(globalCluster.Identifier) == ptr.ToString(suite.globalClusterID) {
			return globalCluster
		}
	}

	return nil
}

func TestRDSGlobalClusterIntegration(t *testing.T) {
	suite.Run(t, new(TestRDSGlobalClusterSuite))
}
