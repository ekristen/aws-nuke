# Writing a New Resource for AWS Nuke

This guide provides step-by-step instructions for adding a new AWS resource to aws-nuke. It covers resource structure, testing strategies, and best practices.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Resource Structure](#resource-structure)
- [Step-by-Step Guide](#step-by-step-guide)
- [Testing](#testing)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)

## Overview

Resources in aws-nuke represent AWS resources that can be listed and deleted. Each resource must implement specific interfaces and follow established patterns for consistency and maintainability.

## Prerequisites

Before creating a new resource, ensure you have:

1. Go 1.21 or later installed
2. AWS SDK v2 knowledge (we use AWS SDK v2, not v1)
3. Familiarity with the AWS service you're implementing
4. golangci-lint installed for linting
5. Read the [CONTRIBUTING.md](CONTRIBUTING.md) guide

## Resource Structure

Every resource consists of:

1. **Resource File** - `resources/<service-name>-<resource-name>.go`
2. **Mock Tests** - `resources/<service-name>-<resource-name>_mock_test.go` (using gomock)
3. **Integration Tests** - `resources/<service-name>-<resource-name>_test.go` (optional but preferred)

### Basic Resource Template

```go
package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/<servicename>"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MyResourceResource = "MyResource"

func init() {
	registry.Register(&registry.Registration{
		Name:     MyResourceResource,
		Scope:    nuke.Account,
		Resource: &MyResource{},
		Lister:   &MyResourceLister{},
	})
}

type MyResourceLister struct{}

func (l *MyResourceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := servicename.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	// List resources using pagination
	params := &servicename.ListMyResourcesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := servicename.NewListMyResourcesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			resources = append(resources, &MyResource{
				svc:  svc,
				Name: item.Name,
			})
		}
	}

	return resources, nil
}

type MyResource struct {
	svc  *servicename.Client
	Name *string
}

func (r *MyResource) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteMyResource(ctx, &servicename.DeleteMyResourceInput{
		Name: r.Name,
	})
	return err
}

func (r *MyResource) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *MyResource) String() string {
	return *r.Name
}
```

## Step-by-Step Guide

### 1. Create the Resource File

Create a new file in the `resources/` directory following the naming convention:
- Format: `<service>-<resource>.go`
- Example: `eks-clusters.go`, `inspector2.go`

### 2. Define the Resource Constant

```go
const EKSClusterResource = "EKSCluster"
```

The constant should match the resource name used in configuration files.

### 3. Implement the Lister

The lister is responsible for discovering all resources of this type in an AWS account.

#### Simple Lister Example

```go
type EKSClusterLister struct{}

func (l *EKSClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := eks.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &eks.ListClustersInput{
		MaxResults: aws.Int32(100),
	}

	paginator := eks.NewListClustersPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			resources = append(resources, &EKSCluster{
				svc:  svc,
				Name: aws.String(cluster),
			})
		}
	}

	return resources, nil
}
```

**Key Points:**
- Always use context for AWS SDK calls
- Use paginators when available to handle large result sets
- Convert AWS SDK types to your resource struct
- Handle errors appropriately

### 4. Implement the Resource Struct

The resource struct holds the data for a single resource instance.

```go
type EKSCluster struct {
	svc        *eks.Client
	Name       *string
	CreatedAt  *time.Time
	Tags       map[string]string
	settings   *libsettings.Setting
	protection *bool
}
```

**Required Fields:**
- Service client (to make deletion calls)
- Identifier fields (name, ID, ARN, etc.)

**Optional Fields:**
- Timestamps (CreatedAt, LastUpdatedTime)
- Tags
- Settings (if the resource supports settings)
- Status information

### 5. Implement Required Methods

Every resource must implement these methods:

#### Remove Method

```go
func (r *EKSCluster) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: r.Name,
	})
	return err
}
```

#### Properties Method

```go
func (r *EKSCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
```

The Properties method is used for filtering. `NewPropertiesFromStruct` automatically extracts all exported fields.

**Special Property Tags:**
- `property:"tagPrefix=resourceType"` - Use a custom prefix for map fields
- `property:"-"` - Exclude a field from properties

#### String Method

```go
func (r *EKSCluster) String() string {
	return *r.Name
}
```

Returns a human-readable identifier for the resource.

### 6. Register the Resource

In the `init()` function, register your resource:

```go
func init() {
	registry.Register(&registry.Registration{
		Name:     EKSClusterResource,
		Scope:    nuke.Account,
		Resource: &EKSCluster{},
		Lister:   &EKSClusterLister{},
	})
}
```

**With Settings:**

```go
func init() {
	registry.Register(&registry.Registration{
		Name:     EKSClusterResource,
		Scope:    nuke.Account,
		Resource: &EKSCluster{},
		Lister:   &EKSClusterLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}
```

### 7. Implement Settings (Optional)

If your resource supports settings:

```go
func (r *EKSCluster) Settings(setting *libsettings.Setting) {
	r.settings = setting
}
```

Use settings in your Remove method:

```go
func (r *EKSCluster) Remove(ctx context.Context) error {
	if ptr.ToBool(r.protection) && r.settings.GetBool("DisableDeletionProtection") {
		// Disable protection first
		_, err := r.svc.UpdateClusterConfig(ctx, &eks.UpdateClusterConfigInput{
			Name:               r.Name,
			DeletionProtection: aws.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: r.Name,
	})
	return err
}
```

### 8. Implement Filter (Optional)

If some resources should be excluded from deletion:

```go
func (r *MyResource) Filter() error {
	// Example: Skip AWS-managed resources
	if strings.HasPrefix(*r.Path, "/aws-service-role/") {
		return fmt.Errorf("cannot delete service-linked roles")
	}
	return nil
}
```

## Testing

Testing is crucial for ensuring your resource works correctly. We use two types of tests:

### 1. Mock Tests (Required)

Mock tests use gomock to simulate AWS API calls without making real requests.

**File naming:** `<resource-name>_mock_test.go`

#### Basic Mock Test Example

```go
package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go-v2/service/myservice"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_myserviceiface"
)

func Test_Mock_MyResource_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_myserviceiface.NewMockMyServiceAPI(ctrl)

	resource := MyResource{
		svc:  mockSvc,
		Name: ptr.String("test-resource"),
	}

	mockSvc.EXPECT().DeleteMyResource(gomock.Eq(&myservice.DeleteMyResourceInput{
		Name: resource.Name,
	})).Return(&myservice.DeleteMyResourceOutput{}, nil)

	err := resource.Remove(context.TODO())
	a.Nil(err)
}
```

#### Testing Properties

```go
func Test_Mock_MyResource_Properties(t *testing.T) {
	a := assert.New(t)

	resource := MyResource{
		Name: ptr.String("test-name"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("test-name", props.Get("Name"))
	a.Equal("test", props.Get("tag:Environment"))
}
```

#### Testing Lister

```go
func Test_Mock_MyResource_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_myserviceiface.NewMockMyServiceAPI(ctrl)

	mockSvc.EXPECT().ListMyResources(gomock.Any()).Return(&myservice.ListMyResourcesOutput{
		Items: []myservice.Item{
			{
				Name: ptr.String("resource-1"),
			},
			{
				Name: ptr.String("resource-2"),
			},
		},
	}, nil)

	lister := MyResourceLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	resource := resources[0].(*MyResource)
	a.Equal("resource-1", *resource.Name)
}
```

### 2. Integration Tests (Recommended)

Integration tests make real AWS API calls and require actual AWS credentials.

**File naming:** `<resource-name>_test.go`

**Build tag:** `//go:build integration`

#### Basic Integration Test Example

```go
//go:build integration

package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/myservice"
)

type TestMyResourceSuite struct {
	suite.Suite
	svc          *myservice.Client
	resourceName *string
}

func (suite *TestMyResourceSuite) SetupSuite() {
	var err error

	suite.resourceName = ptr.String(fmt.Sprintf("aws-nuke-test-%d", time.Now().UnixNano()))

	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		suite.T().Fatalf("failed to load config, %v", err)
	}

	suite.svc = myservice.NewFromConfig(cfg)

	// Create test resource
	_, err = suite.svc.CreateMyResource(ctx, &myservice.CreateMyResourceInput{
		Name: suite.resourceName,
	})
	if err != nil {
		suite.T().Fatalf("failed to create test resource, %v", err)
	}
}

func (suite *TestMyResourceSuite) TearDownSuite() {
	ctx := context.TODO()

	// Clean up test resource
	_, _ = suite.svc.DeleteMyResource(ctx, &myservice.DeleteMyResourceInput{
		Name: suite.resourceName,
	})
}

func (suite *TestMyResourceSuite) TestList() {
	a := assert.New(suite.T())

	lister := MyResourceLister{}
	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Config: &suite.svc.Options().Config,
	})

	a.Nil(err)
	a.Greater(len(resources), 0)
}

func (suite *TestMyResourceSuite) TestRemove() {
	a := assert.New(suite.T())

	resource := MyResource{
		svc:  suite.svc,
		Name: suite.resourceName,
	}

	err := resource.Remove(context.TODO())
	a.Nil(err)
}

func TestMyResourceIntegration(t *testing.T) {
	suite.Run(t, new(TestMyResourceSuite))
}
```

**Running Integration Tests:**

```bash
go test -tags=integration ./resources/...
```

### 3. Test Coverage Best Practices

Your tests should cover:

1. **Successful removal** - Happy path deletion
2. **Properties** - Verify all properties are correctly exposed
3. **Filtering** - If implemented, test filter logic
4. **Error handling** - Test error scenarios
5. **Edge cases** - Empty lists, nil values, etc.
6. **Settings** - If supported, test setting behavior

## Best Practices

### 1. Use AWS SDK v2

Always use AWS SDK v2 (github.com/aws/aws-sdk-go-v2), not v1.

```go
// ✅ Correct
import "github.com/aws/aws-sdk-go-v2/service/eks"

// ❌ Wrong
import "github.com/aws/aws-sdk-go/service/eks"
```

### 2. Handle Pagination

Always use paginators to handle large result sets:

```go
paginator := eks.NewListClustersPaginator(svc, params)

for paginator.HasMorePages() {
	resp, err := paginator.NextPage(ctx)
	if err != nil {
		return nil, err
	}
	// Process results
}
```

### 3. Filter Undeletable Resources

Filter resources that cannot or should not be deleted:

```go
func (r *MyResource) Filter() error {
	// Skip AWS-managed resources
	if r.IsAWSManaged {
		return fmt.Errorf("cannot delete AWS-managed resource")
	}
	
	// Skip resources already being deleted
	if r.Status == "DELETING" || r.Status == "DELETED" {
		return fmt.Errorf("already deleting")
	}
	
	return nil
}
```

### 4. Use Properties, Not String Functions

Prefer the `Properties()` method over `String()` for filtering:

```go
func (r *MyResource) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
```

This automatically exposes all fields for filtering in config files.

### 5. Handle Tags Correctly

For resources with tags:

```go
type MyResource struct {
	svc  *myservice.Client
	Name *string
	Tags map[string]string
}
```

Tags will automatically be available as `tag:KeyName` in filters.

### 6. Implement Settings When Needed

Use settings for optional behavior:

```go
Settings: []string{
	"DisableDeletionProtection",
	"ForceDelete",
}
```

Users can enable these in their config:

```yaml
MyResource:
  - DisableDeletionProtection: true
```

### 7. Handle Dependencies

Some resources must be deleted in a specific order. Use the `DependsOn` method:

```go
func (r *MyResource) DependsOn() []string {
	return []string{
		"DependentResource1",
		"DependentResource2",
	}
}
```

### 8. Error Handling

Return meaningful errors:

```go
func (r *MyResource) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteMyResource(ctx, &myservice.DeleteMyResourceInput{
		Name: r.Name,
	})
	if err != nil {
		// Let the error propagate - aws-nuke will handle retries
		return err
	}
	return nil
}
```

### 9. Import Organization

Follow the import order specified in CONTRIBUTING.md:

```go
import (
	"context"  // 1. Standard library

	"github.com/gotidy/ptr"  // 2. Third party

	"github.com/aws/aws-sdk-go-v2/aws"  // 3. AWS SDK
	"github.com/aws/aws-sdk-go-v2/service/eks"

	"github.com/ekristen/libnuke/pkg/registry"  // 4. libnuke
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"  // 5. Local packages
)
```

## Common Patterns

### Pattern 1: Simple Resource with Direct Deletion

Used for resources with straightforward list and delete operations.

**Example:** Inspector2

```go
func (l *Inspector2Lister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := inspector2.NewFromConfig(*opts.Config)
	
	resp, err := svc.BatchGetAccountStatus(ctx, &inspector2.BatchGetAccountStatusInput{})
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, a := range resp.Accounts {
		if a.State.Status != inspectortypes.StatusDisabled {
			resources = append(resources, &Inspector2{
				svc:       svc,
				AccountID: a.AccountId,
			})
		}
	}

	return resources, nil
}
```

### Pattern 2: Resource with Additional Describe Calls

Used when the list operation doesn't return all needed information.

**Example:** EKS Clusters

```go
for _, cluster := range resp.Clusters {
	dcResp, err := svc.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(cluster),
	})
	if err != nil {
		return nil, err
	}
	
	resources = append(resources, &EKSCluster{
		svc:        svc,
		Name:       aws.String(cluster),
		CreatedAt:  dcResp.Cluster.CreatedAt,
		Tags:       dcResp.Cluster.Tags,
		protection: dcResp.Cluster.DeletionProtection,
	})
}
```

### Pattern 3: Resource with Protection Settings

Used for resources that have deletion protection that needs to be disabled first.

```go
func (r *EKSCluster) Remove(ctx context.Context) error {
	if ptr.ToBool(r.protection) && r.settings.GetBool("DisableDeletionProtection") {
		_, err := r.svc.UpdateClusterConfig(ctx, &eks.UpdateClusterConfigInput{
			Name:               r.Name,
			DeletionProtection: aws.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: r.Name,
	})
	return err
}
```

### Pattern 4: Resource with Custom Property Handling

Used when you need special property formatting or grouping.

```go
type Inspector2 struct {
	svc           *inspector2.Client
	AccountID     *string
	Status        *inspectortypes.Status
	ResourceState map[string]string `property:"tagPrefix=resourceType"`
}
```

### Pattern 5: Resource with Wait Handler

Used for resources with asynchronous deletion.

```go
func (r *MyResource) Remove(ctx context.Context) error {
	resp, err := r.svc.DeleteServiceLinkedRole(ctx, &iam.DeleteServiceLinkedRoleInput{
		RoleName: r.Name,
	})
	if err != nil {
		return err
	}
	
	r.deletionTaskID = resp.DeletionTaskId
	return nil
}

func (r *MyResource) HandleWait(ctx context.Context) error {
	if r.deletionTaskID == nil {
		return nil
	}

	resp, err := r.svc.GetServiceLinkedRoleDeletionStatus(ctx, &iam.GetServiceLinkedRoleDeletionStatusInput{
		DeletionTaskId: r.deletionTaskID,
	})
	if err != nil {
		return err
	}

	switch *resp.Status {
	case "SUCCEEDED":
		return nil
	case "IN_PROGRESS":
		return liberrors.ErrWaitResource("deletion in progress")
	case "FAILED":
		return fmt.Errorf("deletion failed: %s", *resp.Reason.Reason)
	}
	
	return nil
}
```

## Complete Example

Here's a complete example incorporating all best practices:

```go
package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/example"
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ExampleResourceResource = "ExampleResource"

func init() {
	registry.Register(&registry.Registration{
		Name:     ExampleResourceResource,
		Scope:    nuke.Account,
		Resource: &ExampleResource{},
		Lister:   &ExampleResourceLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type ExampleResourceLister struct{}

func (l *ExampleResourceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := example.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &example.ListResourcesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := example.NewListResourcesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Resources {
			resources = append(resources, &ExampleResource{
				svc:        svc,
				Name:       item.Name,
				CreatedAt:  item.CreatedAt,
				Tags:       item.Tags,
				protection: item.DeletionProtection,
			})
		}
	}

	return resources, nil
}

type ExampleResource struct {
	svc        *example.Client
	Name       *string
	CreatedAt  *time.Time
	Tags       map[string]string
	settings   *libsettings.Setting
	protection *bool
}

func (r *ExampleResource) Remove(ctx context.Context) error {
	if ptr.ToBool(r.protection) && r.settings.GetBool("DisableDeletionProtection") {
		_, err := r.svc.UpdateResourceProtection(ctx, &example.UpdateResourceProtectionInput{
			Name:               r.Name,
			DeletionProtection: aws.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteResource(ctx, &example.DeleteResourceInput{
		Name: r.Name,
	})

	return err
}

func (r *ExampleResource) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ExampleResource) String() string {
	return *r.Name
}

func (r *ExampleResource) Settings(setting *libsettings.Setting) {
	r.settings = setting
}
```

## Checklist

Before submitting your resource:

- [ ] Resource file created in `resources/` directory
- [ ] Uses AWS SDK v2
- [ ] Implements required methods: `Remove()`, `Properties()`, `String()`
- [ ] Registered in `init()` function
- [ ] Handles pagination correctly
- [ ] Mock tests created and passing
- [ ] Properties test included
- [ ] Integration tests created (if possible)
- [ ] Follows import order from CONTRIBUTING.md
- [ ] Code passes `golangci-lint run`
- [ ] Code formatted with `go fmt`
- [ ] Filters undeletable resources appropriately
- [ ] Settings implemented if needed
- [ ] Signed commit with conventional commit message

## Resources

- [CONTRIBUTING.md](CONTRIBUTING.md) - General contribution guidelines
- [AWS SDK for Go v2 Documentation](https://aws.github.io/aws-sdk-go-v2/docs/)
- [libnuke Documentation](https://github.com/ekristen/libnuke)
- [Resource Examples](./resources/) - Browse existing resources for patterns

## Getting Help

If you have questions:

1. Check existing resources in the `resources/` directory for similar patterns
2. Review the [GitHub Discussions](https://github.com/ekristen/aws-nuke/discussions)
3. Open an issue if you encounter problems

## Next Steps

After creating your resource:

1. Test thoroughly with both mock and integration tests
2. Create a Pull Request following the guidelines in CONTRIBUTING.md
3. Ensure all GitHub Actions checks pass
4. Respond to any code review feedback

Happy coding!
