//go:build integration

package resources

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// ---------------------------------------------------------------------------
// Base suite — creates the IAM roles shared by all sub-suites
// ---------------------------------------------------------------------------

type TestECSExpressGatewayServiceSuite struct {
	suite.Suite
	ctx           context.Context
	ecsSvc        *ecs.Client
	iamSvc        *iam.Client
	cfg           aws.Config
	execRoleArn   *string
	infraRoleArn  *string
	execRoleName  *string
	infraRoleName *string
}

func (suite *TestECSExpressGatewayServiceSuite) SetupSuite() {
	suite.ctx = context.TODO()

	cfg, err := config.LoadDefaultConfig(suite.ctx, config.WithRegion("us-east-1"))
	if err != nil {
		suite.T().Fatalf("failed to load AWS config: %v", err)
	}

	suite.cfg = cfg
	suite.ecsSvc = ecs.NewFromConfig(cfg)
	suite.iamSvc = iam.NewFromConfig(cfg)

	ts := time.Now().UnixNano()
	execRoleName := fmt.Sprintf("aws-nuke-testing-exec-%d", ts)
	infraRoleName := fmt.Sprintf("aws-nuke-testing-infra-%d", ts)
	suite.execRoleName = aws.String(execRoleName)
	suite.infraRoleName = aws.String(infraRoleName)

	suite.createExecutionRole(execRoleName)
	suite.createInfrastructureRole(infraRoleName)

}

// createExpressGatewayServiceWithRetry calls CreateExpressGatewayService and retries
// on "Cannot assume role" errors caused by IAM eventual consistency. Newly created
// roles are not immediately assumable by ECS; this retries up to 10 times with a
// 3-second pause between attempts rather than using a fixed sleep estimate.
func (suite *TestECSExpressGatewayServiceSuite) createExpressGatewayServiceWithRetry(input *ecs.CreateExpressGatewayServiceInput) (*ecs.CreateExpressGatewayServiceOutput, error) {
	for attempt := 0; attempt < 10; attempt++ {
		out, err := suite.ecsSvc.CreateExpressGatewayService(suite.ctx, input)
		if err == nil {
			return out, nil
		}
		if strings.Contains(err.Error(), "Cannot assume role") ||
			strings.Contains(err.Error(), "not authorized to perform: sts:AssumeRole") ||
			strings.Contains(err.Error(), "AccessDenied") {
			time.Sleep(3 * time.Second)
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("timed out waiting for IAM roles to propagate")
}

func (suite *TestECSExpressGatewayServiceSuite) createExecutionRole(roleName string) {
	trust := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ecs-tasks.amazonaws.com"},"Action":"sts:AssumeRole"}]}`

	role, err := suite.iamSvc.CreateRole(suite.ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trust),
	})
	if err != nil {
		suite.T().Fatalf("failed to create execution role: %v", err)
	}
	suite.execRoleArn = role.Role.Arn

	_, err = suite.iamSvc.AttachRolePolicy(suite.ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	})
	if err != nil {
		suite.T().Fatalf("failed to attach execution role policy: %v", err)
	}
}

func (suite *TestECSExpressGatewayServiceSuite) createInfrastructureRole(roleName string) {
	// ECS Express Gateway assumes the infrastructure role via an internal delegation role
	// (ECSApplicationInfraManagerDelegationRole) rather than directly as the ecs.amazonaws.com
	// service principal. The second statement allows any account's delegation role to assume
	// this role without hardcoding AWS's internal account ID.
	trust := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ecs.amazonaws.com"},"Action":"sts:AssumeRole"},{"Effect":"Allow","Principal":{"AWS":"*"},"Action":"sts:AssumeRole","Condition":{"StringLike":{"aws:PrincipalArn":"arn:aws:iam::*:role/ECSApplicationInfraManagerDelegationRole"}}}]}`

	role, err := suite.iamSvc.CreateRole(suite.ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trust),
	})
	if err != nil {
		suite.T().Fatalf("failed to create infrastructure role: %v", err)
	}
	suite.infraRoleArn = role.Role.Arn

	_, err = suite.iamSvc.PutRolePolicy(suite.ctx, &iam.PutRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String("ECSExpressInfraPolicy"),
		PolicyDocument: aws.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"elasticloadbalancing:*",
					"ec2:Describe*",
					"ec2:CreateSecurityGroup",
					"ec2:DeleteSecurityGroup",
					"ec2:AuthorizeSecurityGroupIngress",
					"ec2:RevokeSecurityGroupIngress",
					"application-autoscaling:*",
					"ecs:*",
					"iam:PassRole",
					"logs:*"
				],
				"Resource": "*"
			}]
		}`),
	})
	if err != nil {
		suite.T().Fatalf("failed to put infrastructure role policy: %v", err)
	}
}

func (suite *TestECSExpressGatewayServiceSuite) TearDownSuite() {
	if suite.execRoleName != nil {
		_, _ = suite.iamSvc.DetachRolePolicy(suite.ctx, &iam.DetachRolePolicyInput{
			RoleName:  suite.execRoleName,
			PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
		})
		_, _ = suite.iamSvc.DeleteRole(suite.ctx, &iam.DeleteRoleInput{
			RoleName: suite.execRoleName,
		})
	}

	if suite.infraRoleName != nil {
		_, _ = suite.iamSvc.DeleteRolePolicy(suite.ctx, &iam.DeleteRolePolicyInput{
			RoleName:   suite.infraRoleName,
			PolicyName: aws.String("ECSExpressInfraPolicy"),
		})
		_, _ = suite.iamSvc.DeleteRole(suite.ctx, &iam.DeleteRoleInput{
			RoleName: suite.infraRoleName,
		})
	}
}

// ---------------------------------------------------------------------------
// Remove suite
// ---------------------------------------------------------------------------

type TestECSExpressGatewayServiceRemoveSuite struct {
	TestECSExpressGatewayServiceSuite
}

func (suite *TestECSExpressGatewayServiceRemoveSuite) TestRemove() {
	svcName := fmt.Sprintf("aws-nuke-testing-%d", time.Now().UnixNano())

	out, err := suite.createExpressGatewayServiceWithRetry(&ecs.CreateExpressGatewayServiceInput{
		ExecutionRoleArn:      suite.execRoleArn,
		InfrastructureRoleArn: suite.infraRoleArn,
		ServiceName:           aws.String(svcName),
		PrimaryContainer: &ecstypes.ExpressGatewayContainer{
			Image: aws.String("public.ecr.aws/docker/library/nginx:latest"),
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to create ECS Express Gateway Service: %v", err)
	}

	r := &ECSExpressGatewayService{
		svc:        suite.ecsSvc,
		ServiceARN: out.Service.ServiceArn,
		Tags:       make(map[string]string),
	}

	err = r.Remove(suite.ctx)
	assert.Nil(suite.T(), err)
}

// ---------------------------------------------------------------------------
// List suite
// ---------------------------------------------------------------------------

type TestECSExpressGatewayServiceListSuite struct {
	TestECSExpressGatewayServiceSuite
	serviceArn *string
}

func (suite *TestECSExpressGatewayServiceListSuite) SetupSuite() {
	suite.TestECSExpressGatewayServiceSuite.SetupSuite()

	svcName := fmt.Sprintf("aws-nuke-testing-%d", time.Now().UnixNano())

	out, err := suite.createExpressGatewayServiceWithRetry(&ecs.CreateExpressGatewayServiceInput{
		ExecutionRoleArn:      suite.execRoleArn,
		InfrastructureRoleArn: suite.infraRoleArn,
		ServiceName:           aws.String(svcName),
		PrimaryContainer: &ecstypes.ExpressGatewayContainer{
			Image: aws.String("public.ecr.aws/docker/library/nginx:latest"),
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to create ECS Express Gateway Service: %v", err)
	}

	suite.serviceArn = out.Service.ServiceArn
}

func (suite *TestECSExpressGatewayServiceListSuite) TearDownSuite() {
	if suite.serviceArn != nil {
		_, _ = suite.ecsSvc.DeleteExpressGatewayService(suite.ctx, &ecs.DeleteExpressGatewayServiceInput{
			ServiceArn: suite.serviceArn,
		})
	}

	suite.TestECSExpressGatewayServiceSuite.TearDownSuite()
}

func (suite *TestECSExpressGatewayServiceListSuite) TestList() {
	lister := &ECSExpressGatewayServiceLister{}
	resources, err := lister.List(suite.ctx, &nuke.ListerOpts{Config: &suite.cfg})
	assert.Nil(suite.T(), err)

	found := false
	for _, r := range resources {
		express := r.(*ECSExpressGatewayService)
		if aws.ToString(express.ServiceARN) == aws.ToString(suite.serviceArn) {
			found = true
			break
		}
	}

	assert.True(suite.T(), found, "created ECS Express Gateway Service %s not found by lister", aws.ToString(suite.serviceArn))
}

// ---------------------------------------------------------------------------
// Runners
// ---------------------------------------------------------------------------

func TestECSExpressGatewayServiceRemove(t *testing.T) {
	suite.Run(t, new(TestECSExpressGatewayServiceRemoveSuite))
}

func TestECSExpressGatewayServiceList(t *testing.T) {
	suite.Run(t, new(TestECSExpressGatewayServiceListSuite))
}
