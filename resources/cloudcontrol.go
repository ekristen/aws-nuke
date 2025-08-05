package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gotidy/ptr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func init() {
	// It is required to manually define Cloud Control API targets, because
	// existing configs that already filter old-style resources could break,
	// because the resource is also available via Cloud Control.
	//
	// To get an overview of available cloud control resource types run this
	// command in the repo root:
	//     go run ./tools/list-cloudcontrol
	//
	// If there's a resource definition for the resource type, then there's no
	// need to define it here as well, you should use the AlternativeResource
	// property of the resource registration, see ecr-public-repository.go
	// for an example.
	RegisterCloudControl("AWS::AppFlow::ConnectorProfile")
	RegisterCloudControl("AWS::AppFlow::Flow")
	RegisterCloudControl("AWS::AppRunner::Service")
	RegisterCloudControl("AWS::ApplicationInsights::Application")
	RegisterCloudControl("AWS::Backup::Framework")
	RegisterCloudControl("AWS::ECR::PullThroughCacheRule")
	RegisterCloudControl("AWS::ECR::RegistryPolicy")
	RegisterCloudControl("AWS::ECR::ReplicationConfiguration")
	RegisterCloudControl("AWS::MWAA::Environment")
	RegisterCloudControl("AWS::Synthetics::Canary")
	RegisterCloudControl("AWS::Timestream::Database")
	RegisterCloudControl("AWS::Timestream::ScheduledQuery")
	RegisterCloudControl("AWS::Timestream::Table")
	RegisterCloudControl("AWS::Transfer::Workflow")
	RegisterCloudControl("AWS::NetworkFirewall::Firewall")
	RegisterCloudControl("AWS::NetworkFirewall::FirewallPolicy")
	RegisterCloudControl("AWS::NetworkFirewall::LoggingConfiguration")
	RegisterCloudControl("AWS::NetworkFirewall::RuleGroup")
}

// describeRateLimit is a rate limiter to avoid throttling when describing resources via the cloud control api.
// AWS does not publish the rate limits for the cloud control api, the rate seems to be 60 reqs/minute, setting to
// 55 and setting no slack to avoid throttling.
var describeRateLimit = ratelimit.New(55, ratelimit.Per(time.Minute), ratelimit.WithoutSlack)

// RegisterCloudControl registers a resource type for the Cloud Control API. This is a unique function that is used
// in two different places. The first place is in the init() function of this file, where it is used to register
// a select subset of Cloud Control API resource types. The second place is in nuke command file, where it is used
// to dynamically register any resource type provided via the `--cloud-control` flag.
func RegisterCloudControl(typeName string) {
	registry.Register(&registry.Registration{
		Name:     typeName,
		Scope:    nuke.Account,
		Resource: &CloudControlResource{},
		Lister: &CloudControlResourceLister{
			TypeName: typeName,
		},
	})
}

type CloudControlResourceLister struct {
	TypeName string

	logger *logrus.Entry
}

func (l *CloudControlResourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	l.logger = opts.Logger.WithField("type-name", l.TypeName)

	svc := cloudcontrolapi.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudcontrolapi.ListResourcesInput{
		TypeName:   ptr.String(l.TypeName),
		MaxResults: ptr.Int64(100),
	}

	if err := svc.ListResourcesPages(params, func(page *cloudcontrolapi.ListResourcesOutput, lastPage bool) bool {
		dt := describeRateLimit.Take()
		l.logger.Debugf("rate limit time: %s", dt)

		for _, desc := range page.ResourceDescriptions {
			identifier := ptr.ToString(desc.Identifier)
			properties, err := l.cloudControlParseProperties(ptr.ToString(desc.Properties))
			if err != nil {
				l.logger.
					WithError(errors.WithStack(err)).
					WithField("identifier", identifier).
					Error("failed to parse cloud control properties")
				continue
			}
			properties = properties.Set("Identifier", identifier)
			resources = append(resources, &CloudControlResource{
				svc:         svc,
				clientToken: uuid.New().String(),
				typeName:    l.TypeName,
				identifier:  identifier,
				properties:  properties,
			})
		}

		return true
	}); err != nil {
		// If a Type is not available in a region we shouldn't throw an error for it.
		var awsError awserr.Error
		if errors.As(err, &awsError) {
			if awsError.Code() == "TypeNotFoundException" {
				return nil, liberrors.ErrSkipRequest(
					"cloudformation type not available in region: " + *opts.Session.Config.Region)
			}
		}

		return nil, err
	}

	return resources, nil
}

func (l *CloudControlResourceLister) cloudControlParseProperties(payload string) (types.Properties, error) {
	// Warning: The implementation of this function is not very straightforward,
	// because the aws-nuke filter functions expect a very rigid structure and
	// the properties from the Cloud Control API are very dynamic.

	properties := types.NewProperties()
	propMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(payload), &propMap)
	if err != nil {
		return properties, err
	}

	for name, value := range propMap {
		switch v := value.(type) {
		case string:
			properties = properties.Set(name, v)
		case []interface{}:
			for _, value2 := range v {
				switch v2 := value2.(type) {
				case string:
					properties.Set(
						fmt.Sprintf("%s.[%q]", name, v2),
						true,
					)
				case map[string]interface{}:
					if len(v2) == 2 && v2["Key"] != nil && v2["Value"] != nil {
						properties.Set(
							fmt.Sprintf("%s.[%q]", name, v2["Key"]),
							v2["Value"],
						)
					} else {
						l.logger.
							WithField("value", fmt.Sprintf("%q", v)).
							Debugf("nested cloud control property type []%T is not supported", value)
					}
				default:
					l.logger.
						WithField("value", fmt.Sprintf("%q", v)).
						Debugf("nested cloud control property type []%T is not supported", value)
				}
			}

		default:
			// We cannot rely on the default handling of
			// properties.Set, because it would fall back to
			// fmt.Sprintf. Since the cloud control properties are
			// nested it would create properties that are not
			// suitable for filtering. Therefore, we have to
			// implemented more sophisticated parsing.
			l.logger.
				WithField("value", fmt.Sprintf("%q", v)).
				Debugf("cloud control property type %T is not supported", v)
		}
	}

	return properties, nil
}

type CloudControlResource struct {
	svc         *cloudcontrolapi.CloudControlApi
	clientToken string
	typeName    string
	identifier  string
	properties  types.Properties
}

func (r *CloudControlResource) String() string {
	return r.identifier
}

func (r *CloudControlResource) Remove(_ context.Context) error {
	_, err := r.svc.DeleteResource(&cloudcontrolapi.DeleteResourceInput{
		ClientToken: &r.clientToken,
		Identifier:  &r.identifier,
		TypeName:    &r.typeName,
	})
	return err
}

func (r *CloudControlResource) Properties() types.Properties {
	return r.properties
}
