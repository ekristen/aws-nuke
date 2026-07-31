package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ssm" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMAssociationResource = "SSMAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSMAssociationResource,
		Scope:    nuke.Account,
		Resource: &SSMAssociation{},
		Lister:   &SSMAssociationLister{},
	})
}

type ssmAssociationAPI interface {
	DeleteAssociation(*ssm.DeleteAssociationInput) (*ssm.DeleteAssociationOutput, error)
	ListAssociations(*ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error)
	ListTagsForResource(*ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error)
}

type SSMAssociationLister struct {
	mockSvc ssmAssociationAPI
}

func (l *SSMAssociationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc ssmAssociationAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = ssm.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	params := &ssm.ListAssociationsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListAssociations(params)
		if err != nil {
			return nil, err
		}

		for _, association := range output.Associations {
			var tags []*ssm.Tag
			if association.AssociationId != nil {
				tagOutput, err := svc.ListTagsForResource(&ssm.ListTagsForResourceInput{
					ResourceId:   association.AssociationId,
					ResourceType: aws.String(ssm.ResourceTypeForTaggingAssociation),
				})
				if err != nil {
					logrus.WithError(err).
						WithField("association_id", aws.StringValue(association.AssociationId)).
						Warn("unable to list tags for SSM association, skipping to avoid incorrect filtering")
					continue
				}
				if tagOutput != nil {
					tags = tagOutput.TagList
				}
			}

			resources = append(resources, &SSMAssociation{
				svc:             svc,
				AssociationID:   association.AssociationId,
				AssociationName: association.AssociationName,
				DocumentName:    association.Name,
				InstanceID:      association.InstanceId,
				Targets:         association.Targets,
				TargetMaps:      association.TargetMaps,
				Tags:            tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMAssociation struct {
	svc ssmAssociationAPI

	AssociationID   *string                `property:"name=AssociationId" description:"The ID of the SSM association"`
	AssociationName *string                `description:"The optional name of the SSM association"`
	DocumentName    *string                `description:"The name of the SSM document applied by the association"`
	InstanceID      *string                `property:"name=InstanceId" description:"The managed node ID for a legacy association"`
	Targets         []*ssm.Target          `description:"The association targets in key=value form"`
	TargetMaps      []map[string][]*string `description:"The association target maps in key=value form"`
	Tags            []*ssm.Tag             `description:"The tags associated with the association"`
}

func (f *SSMAssociation) Remove(_ context.Context) error {
	_, err := f.svc.DeleteAssociation(&ssm.DeleteAssociationInput{
		AssociationId: f.AssociationID,
	})

	return err
}

func (f *SSMAssociation) Properties() types.Properties {
	properties := types.NewPropertiesFromStruct(f)

	targetInstanceIDs := setAssociationTargetProperties(
		properties,
		"Targets",
		"Target",
		associationTargetValues(f.Targets),
	)
	targetInstanceIDs = append(targetInstanceIDs, setAssociationTargetProperties(
		properties,
		"TargetMaps",
		"TargetMap",
		associationTargetMapValues(f.TargetMaps),
	)...)
	if len(targetInstanceIDs) != 0 {
		properties.Set("TargetInstanceIds", strings.Join(targetInstanceIDs, ","))
	}

	return properties
}

func associationTargetValues(targets []*ssm.Target) map[string][]string {
	targetValues := make(map[string][]string)
	for _, target := range targets {
		if target == nil || target.Key == nil {
			continue
		}

		key := strings.TrimSpace(aws.StringValue(target.Key))
		if key == "" {
			continue
		}

		for _, value := range target.Values {
			if value != nil {
				targetValues[key] = append(targetValues[key], aws.StringValue(value))
			}
		}
	}

	return targetValues
}

func associationTargetMapValues(targetMaps []map[string][]*string) map[string][]string {
	targetValues := make(map[string][]string)
	for _, targetMap := range targetMaps {
		for key, values := range targetMap {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}

			for _, value := range values {
				if value != nil {
					targetValues[key] = append(targetValues[key], aws.StringValue(value))
				}
			}
		}
	}

	return targetValues
}

func setAssociationTargetProperties(
	properties types.Properties,
	aggregateProperty string,
	propertyPrefix string,
	targetValues map[string][]string,
) []string {
	targetKeys := make([]string, 0, len(targetValues))
	for key := range targetValues {
		targetKeys = append(targetKeys, key)
	}
	sort.Strings(targetKeys)

	targets := make([]string, 0, len(targetKeys))
	for _, key := range targetKeys {
		values := strings.Join(targetValues[key], ",")
		properties.SetWithPrefix(propertyPrefix, key, values)
		targets = append(targets, fmt.Sprintf("%s=%s", key, values))
	}
	if len(targets) != 0 {
		properties.Set(aggregateProperty, strings.Join(targets, ";"))
	}

	return targetValues["InstanceIds"]
}

func (f *SSMAssociation) String() string {
	return aws.StringValue(f.AssociationID)
}
