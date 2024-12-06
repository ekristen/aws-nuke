package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ResourceGroupGroupResource = "ResourceGroupGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     ResourceGroupGroupResource,
		Scope:    nuke.Account,
		Resource: &ResourceGroupGroup{},
		Lister:   &ResourceGroupGroupLister{},
	})
}

type ResourceGroupGroupLister struct{}

func (l *ResourceGroupGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := resourcegroups.New(opts.Session)
	var resources []resource.Resource

	params := &resourcegroups.ListGroupsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListGroups(params)
		if err != nil {
			return nil, err
		}

		for _, group := range output.GroupIdentifiers {
			tags, err := svc.GetTags(&resourcegroups.GetTagsInput{
				Arn: group.GroupArn,
			})
			if err != nil {
				logrus.WithError(err).Error("unable to get tags for resource group")
			}

			newResource := &ResourceGroupGroup{
				svc:  svc,
				Name: group.GroupName,
			}

			if tags != nil {
				newResource.Tags = tags.Tags
			}

			resources = append(resources, newResource)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ResourceGroupGroup struct {
	svc  *resourcegroups.ResourceGroups
	Name *string
	Tags map[string]*string
}

func (r *ResourceGroupGroup) Filter() error {
	for k, v := range r.Tags {
		if k == "EnableAWSServiceCatalogAppRegistry" && ptr.ToString(v) == "true" {
			return fmt.Errorf("cannot delete AWS managed resource group")
		}
	}

	return nil
}

func (r *ResourceGroupGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ResourceGroupGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteGroup(&resourcegroups.DeleteGroupInput{
		Group: r.Name,
	})

	return err
}

func (r *ResourceGroupGroup) String() string {
	return *r.Name
}
