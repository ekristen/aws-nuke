//go:build integration

package resources

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type TestLambdaEventSourceMappingSuite struct {
	suite.Suite
	ctx          context.Context
	cfg          aws.Config
	lambdaSvc    *lambda.Client
	sqsSvc       *sqs.Client
	iamSvc       *iam.Client
	functionName string
	roleName     string
	queueURL     string
	queueArn     string
}

func (suite *TestLambdaEventSourceMappingSuite) SetupSuite() {
	suite.ctx = context.TODO()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	suite.functionName = fmt.Sprintf("aws-nuke-testing-lambda-%s", suffix)
	suite.roleName = fmt.Sprintf("aws-nuke-testing-role-%s", suffix)

	cfg, err := config.LoadDefaultConfig(suite.ctx, config.WithRegion("us-east-1"))
	if err != nil {
		suite.T().Fatalf("failed to load config: %v", err)
	}
	suite.cfg = cfg

	suite.lambdaSvc = lambda.NewFromConfig(cfg)
	suite.sqsSvc = sqs.NewFromConfig(cfg)
	suite.iamSvc = iam.NewFromConfig(cfg)

	roleArn := suite.createIAMRole(suffix)
	suite.createSQSQueue(suffix)
	suite.createLambdaFunction(roleArn)
}

func (suite *TestLambdaEventSourceMappingSuite) createIAMRole(suffix string) string {
	roleResp, err := suite.iamSvc.CreateRole(suite.ctx, &iam.CreateRoleInput{
		RoleName: aws.String(suite.roleName),
		AssumeRolePolicyDocument: aws.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"Service": "lambda.amazonaws.com"},
				"Action": "sts:AssumeRole"
			}]
		}`),
	})
	if err != nil {
		suite.T().Fatalf("failed to create IAM role: %v", err)
	}

	_, err = suite.iamSvc.AttachRolePolicy(suite.ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(suite.roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"),
	})
	if err != nil {
		suite.T().Fatalf("failed to attach execution policy: %v", err)
	}

	return *roleResp.Role.Arn
}

func (suite *TestLambdaEventSourceMappingSuite) createSQSQueue(suffix string) {
	queueResp, err := suite.sqsSvc.CreateQueue(suite.ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(fmt.Sprintf("aws-nuke-testing-sqs-%s", suffix)),
	})
	if err != nil {
		suite.T().Fatalf("failed to create SQS queue: %v", err)
	}
	suite.queueURL = *queueResp.QueueUrl

	attrResp, err := suite.sqsSvc.GetQueueAttributes(suite.ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       queueResp.QueueUrl,
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
	})
	if err != nil {
		suite.T().Fatalf("failed to get queue ARN: %v", err)
	}
	suite.queueArn = attrResp.Attributes[string(sqstypes.QueueAttributeNameQueueArn)]
}

func (suite *TestLambdaEventSourceMappingSuite) createLambdaFunction(roleArn string) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("index.py")
	if err != nil {
		suite.T().Fatalf("failed to create zip entry: %v", err)
	}
	_, err = f.Write([]byte("def handler(event, context): return {}"))
	if err != nil {
		suite.T().Fatalf("failed to write zip content: %v", err)
	}
	if err = zw.Close(); err != nil {
		suite.T().Fatalf("failed to close zip: %v", err)
	}

	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String(suite.functionName),
		Role:         aws.String(roleArn),
		Runtime:      lambdatypes.RuntimePython314,
		Handler:      aws.String("index.handler"),
		Code:         &lambdatypes.FunctionCode{ZipFile: buf.Bytes()},
	}

	// Retry CreateFunction with backoff until the IAM role propagates.
	deadline := time.Now().Add(2 * time.Minute)
	delay := 2 * time.Second
	for {
		_, err = suite.lambdaSvc.CreateFunction(suite.ctx, input)
		if err == nil {
			break
		}
		// Lambda returns InvalidParameterValueException when the IAM role has not
		// yet propagated globally; retry with backoff until it does.
		var invalidParam *lambdatypes.InvalidParameterValueException
		if !errors.As(err, &invalidParam) || time.Now().After(deadline) {
			suite.T().Fatalf("failed to create Lambda function: %v", err)
		}
		time.Sleep(delay)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	waiter := lambda.NewFunctionActiveV2Waiter(suite.lambdaSvc)
	if err = waiter.Wait(suite.ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(suite.functionName),
	}, 2*time.Minute); err != nil {
		suite.T().Fatalf("Lambda function did not become active: %v", err)
	}
}

func (suite *TestLambdaEventSourceMappingSuite) TearDownSuite() {
	// Best-effort cleanup of supporting resources
	_, _ = suite.lambdaSvc.DeleteFunction(suite.ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(suite.functionName),
	})

	_, _ = suite.iamSvc.DetachRolePolicy(suite.ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(suite.roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"),
	})

	_, _ = suite.iamSvc.DeleteRole(suite.ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(suite.roleName),
	})

	_, _ = suite.sqsSvc.DeleteQueue(suite.ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(suite.queueURL),
	})
}

func (suite *TestLambdaEventSourceMappingSuite) Test_List() {
	// Create a mapping with a tag so we can verify tag retrieval
	mappingResp, err := suite.lambdaSvc.CreateEventSourceMapping(suite.ctx, &lambda.CreateEventSourceMappingInput{
		EventSourceArn: aws.String(suite.queueArn),
		FunctionName:   aws.String(suite.functionName),
		Enabled:        aws.Bool(false),
		Tags:           map[string]string{"aws-nuke-test": "true"},
	})
	assert.NoError(suite.T(), err)

	defer func() {
		_, _ = suite.lambdaSvc.DeleteEventSourceMapping(suite.ctx, &lambda.DeleteEventSourceMappingInput{
			UUID: mappingResp.UUID,
		})
	}()

	lister := &LambdaEventSourceMappingLister{}
	resources, err := lister.List(suite.ctx, &nuke.ListerOpts{Config: &suite.cfg})
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(resources), 0)

	var found *LambdaEventSourceMapping
	for _, r := range resources {
		mapping := r.(*LambdaEventSourceMapping)
		if *mapping.UUID == *mappingResp.UUID {
			found = mapping
			break
		}
	}
	assert.NotNil(suite.T(), found, "expected to find the created event source mapping in list results")
	if found != nil {
		assert.Equal(suite.T(), "true", found.Tags["aws-nuke-test"], "expected tag to be present on listed mapping")
	}
}

func (suite *TestLambdaEventSourceMappingSuite) Test_Remove() {
	// Create a mapping to delete
	mappingResp, err := suite.lambdaSvc.CreateEventSourceMapping(suite.ctx, &lambda.CreateEventSourceMappingInput{
		EventSourceArn: aws.String(suite.queueArn),
		FunctionName:   aws.String(suite.functionName),
		Enabled:        aws.Bool(false),
	})
	assert.NoError(suite.T(), err)

	r := &LambdaEventSourceMapping{
		svc:  suite.lambdaSvc,
		UUID: mappingResp.UUID,
	}

	err = r.Remove(suite.ctx)
	assert.NoError(suite.T(), err)

	// Verify deletion: the mapping should no longer be found
	out, err := suite.lambdaSvc.ListEventSourceMappings(suite.ctx, &lambda.ListEventSourceMappingsInput{
		EventSourceArn: aws.String(suite.queueArn),
		FunctionName:   aws.String(suite.functionName),
	})
	assert.NoError(suite.T(), err)
	for _, m := range out.EventSourceMappings {
		assert.NotEqual(suite.T(), *mappingResp.UUID, *m.UUID)
	}
}

func TestLambdaEventSourceMappingIntegration(t *testing.T) {
	suite.Run(t, new(TestLambdaEventSourceMappingSuite))
}
