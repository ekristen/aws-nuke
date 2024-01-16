# Testing

This is not a lot of test coverage around the resources themselves. This is due to the cost of running the tests. However,
[libnuke](https://github.com/ekristen/libnuke) is extensively tested for functionality to ensure a smooth experience.

Generally speaking, the tests are split into two categories:

1. Tool Testing
2. Resource Testing

Furthermore, for resource testing, these are broken down into two additional categories:

1. Mock Tests
2. Integration Tests

## Tool Testing

These are unit tests written against non resource focused code inside the `pkg/` directory.

## Resource Testing

These are unit tests written against the resources in the `resources/` directory.

### Mock Tests

These are tests where the AWS API calls are mocked out. This is done to ensure that the code is working as expected.
Currently, there are only two services mocked out for testing, IAM and CloudFormation.

#### Adding Additional Mocks

To add another service to be mocked out, you will need to do the following:

1. Identify the service in the AWS SDK for Go
2. Create a new file in the `resources/` directory called `service_mock_test.go`
3. Add the following code to the file: (replace `<service>` with actual service name)
    ```go
    //go:generate ../mocks/generate_mocks.sh <service> <service>iface
    package resources
    
    // Note: empty on purpose, this file exist purely to generate mocks for the <service> service
    ```
4. Run `make generate` to generate the mocks
5. Add tests to the `resources/<service>_mock_test.go` file.
6. Run `make test` to ensure the tests pass
7. Submit a PR with the changes

### Integration Tests

These are tests where the AWS API calls are called directly and tested against a live AWS account. These tests are
behind a build flag (`-tags=integration`), so they are not run by default. To run these tests, you will need to run the following:

```bash
make test-integration
```

#### Adding Additional Integration Tests

To add another integration test, you will need to do the following:

1. Create a new file in the `resources/` directory called `<resource>_test.go`
2. Add the following code to the file: (replace `<resource>` with actual resource name)
    ```go
    //go:build integration
    
    package resources
    
    import (
        "testing"
    
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/<resource>"
    )
    
    func Test_ExampleResource_Remove(t *testing.T) {
        // 1. write code to create resource in AWS using golang sdk
        // 2. stub the resource struct out that is defined in <resource>.go file
        // 3. call the Remove() function
        // 4. assert that the resource was removed
    }
   ```
3. Run `make test-integration` to ensure the tests pass
4. Submit a PR with the changes
