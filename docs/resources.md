# Resources

Resources are the core of the tool, they are what is used to list and remove resources from AWS. The resources are
broken down into separate files. 

When creating a resource there's the base resource type, then there's the `Lister` type that returns a list of resources
that it discovers. Those resources are then filtered by any filtering criteria on the resource itself.

## Anatomy of a Resource

The anatomy of a resource is fairly simple, it's broken down into a few parts:

- `Resource` - This is the base resource type that is used to define the resource.
- `Lister` - This is the type that is used to list the resources.

### Resource

The resource must have the `func Remove() error` method defined on it, this is what is used to remove the resource.

It can optionally have the following methods defined:

- `func Filter() error` - This is used to pre-filter resources, usually based on internal criteria, like system defaults.
- `func String() string` - This is used to print the resource in a human-readable format.
- `func Properties() types.Properties` - This is used to print the resource in a human-readable format.

```go
package resources

import (
    "context"
    
    "github.com/ekristen/libnuke/pkg/resource"
    "github.com/ekristen/libnuke/pkg/types"

    "github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type ExampleResource struct {
    ID *string
}

func (r *ExampleResource) Remove(_ context.Context) error {
    // remove the resource, an error will put the resource in failed state
    // resources in failed state are retried a number of times
    return nil
}

func (r *ExampleResource) Filter() error {
    // filter the resource, this is useful for built-in resources that cannot
    // be removed, like an AWS managed resource, return an error here to filter
    // it before it even gets to the user supplied filters.
    return nil
}

func (r *ExampleResource) String() string {
    // return a string representation of the resource, this is legacy, but still
    // used for a number of reasons.
    return *r.ID
}
```

## Lister

The lister must have the `func List(ctx context.Context, o interface{}) ([]resource.Resource, error)` method defined on it.

```go
package resources

import (
	"context"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type ExampleResourceLister struct{}

func (l *ExampleResourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
    opts := o.(*nuke.ListerOpts)

    var resources []resource.Resource
    
    // list the resources and add to resources slice

    return resources, nil
}
```

### Example

```go
package resources

import (
	"context"
	
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type ExampleResourceLister struct{}

func (l *ExampleResourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var resources []resource.Resource
	
	// list the resources and add to resources slice

	return resources, nil
}

// -----------------------------------------------------------------------------

type ExampleResource struct {
	ID *string
}

func (r *ExampleResource) Remove(_ context.Context) error {
	// remove the resource, an error will put the resource in failed state
	// resources in failed state are retried a number of times
	return nil
}

func (r *ExampleResource) Filter() error {
	// filter the resource, this is useful for built-in resources that cannot
	// be removed, like an AWS managed resource, return an error here to filter
	// it before it even gets to the user supplied filters.
	return nil
}

func (r *ExampleResource) String() string {
	// return a string representation of the resource, this is legacy, but still
	// used for a number of reasons.
	return *r.ID
}

func (r *ExampleResource) Properties() types.Properties {
	// return a properties representation of the resource
	props := types.NewProperties()
	props.Set("ID", r.ID)
	return props
}
```

## Creating a new resource

Creating a new resources is fairly straightforward and a template is provided for you, along with a tool to help you
generate the boilerplate code.

Currently, the code is generated using a tool that is located in `tools/create-resource/main.go` and can be run like so:

!!! note
    At present, the tool does not check if the service or the resource type is valid, this is purely a helper tool to
    generate the boilerplate code.

```bash
go run tools/create-resource/main.go <service> <resource-type>
```

This will output the boilerplate code to stdout, so you can copy and paste it into the appropriate file or you can
redirect to a file like so:

```bash
go run tools/create-resource/main.go <service> <resource-type> > resources/<resource-type>.go
```

## Converting a resource for self documenting

To convert a resource for self documenting, you need to do the following:

- Write a test for the resource to verify its current properties
- Add the `Resource` field to the registration struct
- Capitalize the first letter of the field names
- Match the field names to the property names that are defined in `Properties()` method
- Switch to `NewPropertiesFromStruct` method in `Properties()` method
- Run the tests to verify the properties are still correct
- Run the `generate-resource-docs` command to generate the documentation
- Commit your changes and open a pull request

### Example

**Before**

```go
package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SNSTopicResource = "SNSTopic"

func init() {
	registry.Register(&registry.Registration{
		Name:   SNSTopicResource,
		Scope:  nuke.Account,
		Lister: &SNSTopicLister{},
	})
}

type SNSTopic struct {
	svc  *sns.SNS
	id   *string
	tags []*sns.Tag
}

func (r *SNSTopic) Remove(_ context.Context) error {
	_, err := r.svc.DeleteTopic(&sns.DeleteTopicInput{
		TopicArn: r.id,
	})
	return err
}

func (r *SNSTopic) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.Set("TopicARN", r.id)

	return properties
}

func (r *SNSTopic) String() string {
	return fmt.Sprintf("TopicARN: %s", *r.id)
}

type SNSTopicLister struct{}

func (l *SNSTopicLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	return nil, nil
}
```

**After**

```go
package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SNSTopicResource = "SNSTopic"

func init() {
	registry.Register(&registry.Registration{
		Name:     SNSTopicResource,
		Scope:    nuke.Account,
		Resource: &SNSTopic{},
		Lister:   &SNSTopicLister{},
	})
}

type SNSTopic struct {
	svc      *sns.SNS
	TopicARN *string
	Tags     []*sns.Tag
}

func (r *SNSTopic) Remove(_ context.Context) error {
	_, err := r.svc.DeleteTopic(&sns.DeleteTopicInput{
		TopicArn: r.TopicARN,
	})
	return err
}

func (r *SNSTopic) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SNSTopic) String() string {
	return fmt.Sprintf("TopicARN: %s", *r.TopicARN)
}

type SNSTopicLister struct{}

func (l *SNSTopicLister) List(_ context.Context, _ interface{}) ([]resource.Resource, error) {
	return nil, nil
}
```
