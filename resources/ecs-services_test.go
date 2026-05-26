//go:build integration

package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// ---------------------------------------------------------------------------
// Base suite — creates a cluster and task definition shared by all sub-suites
// ---------------------------------------------------------------------------

type TestECSServiceSuite struct {
	suite.Suite
	ctx        context.Context
	svc        *ecs.Client
	cfg        aws.Config
	clusterArn *string
	taskDefArn *string
}

func (suite *TestECSServiceSuite) SetupSuite() {
	suite.ctx = context.TODO()

	cfg, err := config.LoadDefaultConfig(suite.ctx, config.WithRegion("us-east-1"))
	if err != nil {
		suite.T().Fatalf("failed to load AWS config: %v", err)
	}

	suite.cfg = cfg
	suite.svc = ecs.NewFromConfig(cfg)

	clusterName := fmt.Sprintf("aws-nuke-testing-%d", time.Now().UnixNano())
	clusterOut, err := suite.svc.CreateCluster(suite.ctx, &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
	})
	if err != nil {
		suite.T().Fatalf("failed to create ECS cluster: %v", err)
	}

	suite.clusterArn = clusterOut.Cluster.ClusterArn

	taskFamily := fmt.Sprintf("aws-nuke-testing-%d", time.Now().UnixNano())
	taskDefOut, err := suite.svc.RegisterTaskDefinition(suite.ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(taskFamily),
		ContainerDefinitions: []ecstypes.ContainerDefinition{
			{
				Name:   aws.String("test"),
				Image:  aws.String("public.ecr.aws/docker/library/alpine:latest"),
				Memory: aws.Int32(128),
			},
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to register task definition: %v", err)
	}

	suite.taskDefArn = taskDefOut.TaskDefinition.TaskDefinitionArn
}

func (suite *TestECSServiceSuite) TearDownSuite() {
	if suite.taskDefArn != nil {
		_, _ = suite.svc.DeregisterTaskDefinition(suite.ctx, &ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: suite.taskDefArn,
		})
		_, _ = suite.svc.DeleteTaskDefinitions(suite.ctx, &ecs.DeleteTaskDefinitionsInput{
			TaskDefinitions: []string{aws.ToString(suite.taskDefArn)},
		})
	}

	if suite.clusterArn != nil {
		_, _ = suite.svc.DeleteCluster(suite.ctx, &ecs.DeleteClusterInput{
			Cluster: suite.clusterArn,
		})
	}
}

// ---------------------------------------------------------------------------
// Remove suite
// ---------------------------------------------------------------------------

type TestECSServiceRemoveSuite struct {
	TestECSServiceSuite
}

func (suite *TestECSServiceRemoveSuite) TestRemove() {
	serviceName := fmt.Sprintf("aws-nuke-testing-%d", time.Now().UnixNano())
	resp, err := suite.svc.CreateService(suite.ctx, &ecs.CreateServiceInput{
		Cluster:        suite.clusterArn,
		ServiceName:    aws.String(serviceName),
		TaskDefinition: suite.taskDefArn,
		DesiredCount:   aws.Int32(0),
	})
	if err != nil {
		suite.T().Fatalf("failed to create ECS service: %v", err)
	}

	ecsService := &ECSService{
		svc:        suite.svc,
		ServiceARN: resp.Service.ServiceArn,
		ClusterARN: suite.clusterArn,
		Tags:       make(map[string]string),
	}

	err = ecsService.Remove(suite.ctx)
	assert.Nil(suite.T(), err)
}

// ---------------------------------------------------------------------------
// List suite
// ---------------------------------------------------------------------------

type TestECSServiceListSuite struct {
	TestECSServiceSuite
	serviceArn *string
}

func (suite *TestECSServiceListSuite) SetupSuite() {
	suite.TestECSServiceSuite.SetupSuite()

	serviceName := fmt.Sprintf("aws-nuke-testing-%d", time.Now().UnixNano())
	resp, err := suite.svc.CreateService(suite.ctx, &ecs.CreateServiceInput{
		Cluster:        suite.clusterArn,
		ServiceName:    aws.String(serviceName),
		TaskDefinition: suite.taskDefArn,
		DesiredCount:   aws.Int32(0),
	})
	if err != nil {
		suite.T().Fatalf("failed to create ECS service: %v", err)
	}

	suite.serviceArn = resp.Service.ServiceArn
}

func (suite *TestECSServiceListSuite) TearDownSuite() {
	if suite.serviceArn != nil {
		_, _ = suite.svc.DeleteService(suite.ctx, &ecs.DeleteServiceInput{
			Cluster: suite.clusterArn,
			Service: suite.serviceArn,
			Force:   aws.Bool(true),
		})
	}

	suite.TestECSServiceSuite.TearDownSuite()
}

func (suite *TestECSServiceListSuite) TestList() {
	lister := &ECSServiceLister{}
	resources, err := lister.List(suite.ctx, &nuke.ListerOpts{Config: &suite.cfg})
	assert.Nil(suite.T(), err)

	found := false
	for _, r := range resources {
		ecsService := r.(*ECSService)
		if aws.ToString(ecsService.ServiceARN) == aws.ToString(suite.serviceArn) {
			found = true
			break
		}
	}

	assert.True(suite.T(), found, "created ECS service %s not found by lister", aws.ToString(suite.serviceArn))
}

// ---------------------------------------------------------------------------
// Runners
// ---------------------------------------------------------------------------

func TestECSServiceRemove(t *testing.T) {
	suite.Run(t, new(TestECSServiceRemoveSuite))
}

func TestECSServiceList(t *testing.T) {
	suite.Run(t, new(TestECSServiceListSuite))
}
