package resources

import (
	"context"

	"github.com/ekristen/libnuke/pkg/featureflag"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSInstanceResource = "RDSInstance"

func init() {
	resource.Register(resource.Registration{
		Name:   RDSInstanceResource,
		Scope:  nuke.Account,
		Lister: &RDSInstanceLister{},
	})
}

type RDSInstance struct {
	svc      *rds.RDS
	instance *rds.DBInstance
	tags     []*rds.Tag

	featureFlags *featureflag.FeatureFlags
}

type RDSInstanceLister struct{}

func (l *RDSInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := rds.New(opts.Session)

	params := &rds.DescribeDBInstancesInput{}
	resp, err := svc.DescribeDBInstances(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, instance := range resp.DBInstances {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: instance.DBInstanceArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSInstance{
			svc:      svc,
			instance: instance,
			tags:     tags.TagList,
		})
	}

	return resources, nil
}

func (i *RDSInstance) FeatureFlags(ff *featureflag.FeatureFlags) {
	i.featureFlags = ff
}

func (i *RDSInstance) Remove(_ context.Context) error {
	ffddpRDSInstance, err := i.featureFlags.Get("DisableDeletionProtection_RDSInstance")
	if err != nil {
		return err
	}

	if aws.BoolValue(i.instance.DeletionProtection) && ffddpRDSInstance.Enabled() {
		modifyParams := &rds.ModifyDBInstanceInput{
			DBInstanceIdentifier: i.instance.DBInstanceIdentifier,
			DeletionProtection:   aws.Bool(false),
		}
		if _, err := i.svc.ModifyDBInstance(modifyParams); err != nil {
			return err
		}
	}

	params := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: i.instance.DBInstanceIdentifier,
		SkipFinalSnapshot:    aws.Bool(true),
	}

	if _, err := i.svc.DeleteDBInstance(params); err != nil {
		return err
	}

	return nil
}

func (i *RDSInstance) Properties() types.Properties {
	properties := types.NewProperties().
		Set("Identifier", i.instance.DBInstanceIdentifier).
		Set("DeletionProtection", i.instance.DeletionProtection).
		Set("AvailabilityZone", i.instance.AvailabilityZone).
		Set("InstanceClass", i.instance.DBInstanceClass).
		Set("Engine", i.instance.Engine).
		Set("EngineVersion", i.instance.EngineVersion).
		Set("MultiAZ", i.instance.MultiAZ).
		Set("PubliclyAccessible", i.instance.PubliclyAccessible)

	if i.instance.InstanceCreateTime != nil {
		properties.Set("InstanceCreateTime", i.instance.InstanceCreateTime.Format(time.RFC3339))
	}

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (i *RDSInstance) String() string {
	return aws.StringValue(i.instance.DBInstanceIdentifier)
}
