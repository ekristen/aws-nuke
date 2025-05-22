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
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
)

var ctx = context.TODO()
var fileSystemName string

type TestEFSSuite struct {
	suite.Suite
	svc *efs.Client
}

func (suite *TestEFSSuite) SetupSuite() {
	fileSystemName = fmt.Sprintf("aws-nuke-testing-efs-%d", time.Now().UnixNano())

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		suite.T().Fatalf("failed to create config, %v", err)
	}

	suite.svc = efs.NewFromConfig(cfg)
}

func (suite *TestEFSSuite) TearDownSuite() {

}

type BasicEFSTestSuite struct {
	TestEFSSuite
}

func (suite *BasicEFSTestSuite) Test() {

	resp, err := suite.svc.CreateFileSystem(ctx, &efs.CreateFileSystemInput{
		Tags: []types.Tag{{Key: aws.String("Name"), Value: aws.String(fileSystemName)}},
	})

	if err != nil {
		assert.Nil(suite.T(), err)
	}

	fs := &EFSFileSystem{
		svc: suite.svc,
		id:  *resp.FileSystemId,
	}

	assert.Equal(suite.T(), fileSystemName, *resp.Name)

	err = fs.Remove(ctx)
	assert.Nil(suite.T(), err)
}

func TestEFSFilesystem(t *testing.T) {
	suite.Run(t, new(BasicEFSTestSuite))
}
