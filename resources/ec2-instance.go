package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2InstanceResource = "EC2Instance"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2InstanceResource,
		Scope:  nuke.Account,
		Lister: &EC2InstanceLister{},
	})
}

type EC2InstanceLister struct{}

func (l *EC2InstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	params := &ec2.DescribeInstancesInput{}
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.DescribeInstances(params)
		if err != nil {
			return nil, err
		}

		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				resources = append(resources, &EC2Instance{
					svc:      svc,
					instance: instance,
				})
			}
		}

		if resp.NextToken == nil {
			break
		}

		params = &ec2.DescribeInstancesInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type EC2Instance struct {
	svc      *ec2.EC2
	instance *ec2.Instance

	settings *libsettings.Setting
}

func (i *EC2Instance) Settings(setting *libsettings.Setting) {
	i.settings = setting
}

func (i *EC2Instance) Filter() error {
	if *i.instance.State.Name == "terminated" {
		return fmt.Errorf("already terminated")
	}
	return nil
}

func (i *EC2Instance) Remove(_ context.Context) error {
	deleteTagsParams := &ec2.DeleteTagsInput{
		Resources: []*string{i.instance.InstanceId},
	}
	if _, err := i.svc.DeleteTags(deleteTagsParams); err != nil {
		return err
	}

	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{i.instance.InstanceId},
	}

	if _, err := i.svc.TerminateInstances(params); err != nil {
		var awsErr awserr.Error
		ok := errors.As(err, &awsErr)

		// Check for Termination Protection, disable it, and try termination again.
		if ok && awsErr.Code() == awsutil.ErrCodeOperationNotPermitted &&
			awsErr.Message() == "The instance '"+*i.instance.InstanceId+"' may not be "+
				"terminated. Modify its 'disableApiTermination' instance attribute and "+
				"try again." && i.settings.GetBool("DisableDeletionProtection") {
			termErr := i.DisableTerminationProtection()
			if termErr != nil {
				return termErr
			}
			_, err = i.svc.TerminateInstances(params)
			// If we still get an error, we'll check for type and let the next routine
			// handle it.
			if err != nil {
				ok = errors.As(err, &awsErr)
			}
		}

		// Check for Stop Protection, disable it, and try termination again.
		if ok && awsErr.Code() == "OperationNotPermitted" &&
			awsErr.Message() == "The instance '"+*i.instance.InstanceId+"' may not be "+
				"terminated. Modify its 'disableApiStop' instance attribute and try "+
				"again." && i.settings.GetBool("DisableStopProtection") {
			stopErr := i.DisableStopProtection()
			if stopErr != nil {
				return stopErr
			}
			_, err = i.svc.TerminateInstances(params)
		}

		// If we still have an error at this point, we'll return it.
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *EC2Instance) DisableStopProtection() error {
	params := &ec2.ModifyInstanceAttributeInput{
		InstanceId: i.instance.InstanceId,
		DisableApiStop: &ec2.AttributeBooleanValue{
			Value: aws.Bool(false),
		},
	}
	_, err := i.svc.ModifyInstanceAttribute(params)
	if err != nil {
		return err
	}
	return nil
}

func (i *EC2Instance) DisableTerminationProtection() error {
	params := &ec2.ModifyInstanceAttributeInput{
		InstanceId: i.instance.InstanceId,
		DisableApiTermination: &ec2.AttributeBooleanValue{
			Value: aws.Bool(false),
		},
	}
	_, err := i.svc.ModifyInstanceAttribute(params)
	if err != nil {
		return err
	}
	return nil
}

func (i *EC2Instance) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Identifier", i.instance.InstanceId)
	properties.Set("ImageIdentifier", i.instance.ImageId)
	properties.Set("InstanceState", i.instance.State.Name)
	properties.Set("InstanceType", i.instance.InstanceType)
	properties.Set("LaunchTime", i.instance.LaunchTime.Format(time.RFC3339))

	for _, tagValue := range i.instance.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (i *EC2Instance) String() string {
	return *i.instance.InstanceId
}
