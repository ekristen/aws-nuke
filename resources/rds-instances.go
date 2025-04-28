package resources

import (
	"context"
	"errors"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"

	liberror "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const RDSInstanceResource = "RDSInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     RDSInstanceResource,
		Scope:    nuke.Account,
		Resource: &RDSInstance{},
		Lister:   &RDSInstanceLister{},
		Settings: []string{
			"DisableDeletionProtection",
			"StartClusterToDelete",
		},
	})
}

type RDSInstance struct {
	svc      *rds.RDS
	instance *rds.DBInstance
	tags     []*rds.Tag

	settings *libsettings.Setting
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
		// Note: NeptuneInstance handles Neptune instances
		if ptr.ToString(instance.Engine) == "neptune" {
			opts.Logger.Debug("skipping neptune instance, it is handled by NeptuneInstance")
			continue
		}

		if ptr.ToString(instance.Engine) == "docdb" {
			opts.Logger.Debug("skipping docdb instance, it is handled by DocDBInstance")
			continue
		}

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

func (i *RDSInstance) Settings(settings *libsettings.Setting) {
	i.settings = settings
}

func (i *RDSInstance) Remove(_ context.Context) error {
	status, err := i.getDBInstanceStatus()
	if err != nil {
		return err
	}
	if status == awsutil.StateDeleting {
		return nil
	}

	// You can't delete an instance that is part of a cluster in the stopped state.
	// If the setting is enabled, start the cluster before deleting the instance.
	if i.settings.GetBool("StartClusterToDelete") {
		status, err = i.getDBClusterStatus()
		if err != nil {
			return err
		}
		switch status {
		case "stopped", "inaccessible-encryption-credentials-recoverable":
			_, err := i.svc.StartDBCluster(&rds.StartDBClusterInput{
				DBClusterIdentifier: i.instance.DBClusterIdentifier,
			})
			return err
		case "starting":
			return nil
		}
	}

	return i.deleteDBInstance()
}

func (i *RDSInstance) deleteDBInstance() error {
	if aws.BoolValue(i.instance.DeletionProtection) && i.settings.GetBool("DisableDeletionProtection") {
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

func (i *RDSInstance) getDBInstanceStatus() (string, error) {
	resp, err := i.svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: i.instance.DBInstanceIdentifier,
	})
	if err != nil {
		return "", err
	}
	if len(resp.DBInstances) == 0 {
		return "", nil
	}

	return ptr.ToString(resp.DBInstances[0].DBInstanceStatus), nil
}

func (i *RDSInstance) getDBClusterStatus() (string, error) {
	if i.instance.DBClusterIdentifier == nil {
		// No cluster associated with this instance
		return "", nil
	}

	cluster, err := i.svc.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: i.instance.DBClusterIdentifier,
	})
	if err != nil {
		return "", err
	}
	if len(cluster.DBClusters) == 0 {
		return "", nil
	}

	return ptr.ToString(cluster.DBClusters[0].Status), nil
}

func (i *RDSInstance) Properties() types.Properties {
	properties := types.NewProperties().
		Set("Identifier", i.instance.DBInstanceIdentifier).
		Set("ClusterIdentifier", i.instance.DBClusterIdentifier).
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

func (i *RDSInstance) HandleWait(ctx context.Context) error {
	status, err := i.getDBInstanceStatus()
	if err != nil {
		var awsErr awserr.Error
		ok := errors.As(err, &awsErr)
		if ok && awsErr.Code() == "DBInstanceNotFound" {
			return nil
		}

		return err
	}
	if status == awsutil.StateDeleting {
		return liberror.ErrWaitResource("waiting for instance to delete")
	}

	if i.settings.GetBool("StartClusterToDelete") {
		status, err = i.getDBClusterStatus()
		if err != nil {
			return err
		}
		switch status {
		case "starting":
			return liberror.ErrWaitResource("waiting for cluster to start")
		case "available":
			return i.deleteDBInstance()
		}
	}

	return nil
}
