package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

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
		Name:     EC2InstanceResource,
		Scope:    nuke.Account,
		Resource: &EC2Instance{},
		Lister:   &EC2InstanceLister{},
		Settings: []string{
			"DisableDeletionProtection",
			"DisableStopProtection",
		},
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
					svc:          svc,
					ID:           instance.InstanceId,
					ImageID:      instance.ImageId,
					State:        instance.State.Name,
					InstanceType: instance.InstanceType,
					LaunchTime:   instance.LaunchTime,
					Tags:         instance.Tags,
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
	settings *libsettings.Setting

	ID           *string    `property:"name=Identifier" description:"The instance ID (e.g. i-1234567890abcdef0)"`
	ImageID      *string    `property:"name=ImageIdentifier" description:"The ID of the AMI used to launch the instance"`
	State        *string    `property:"name=InstanceState" description:"The current state of the instance"`
	InstanceType *string    `description:"The instance type (e.g. t2.micro)"`
	LaunchTime   *time.Time `description:"The time the instance was launched"`
	Tags         []*ec2.Tag `description:"The tags associated with the instance"`
}

func (i *EC2Instance) Settings(setting *libsettings.Setting) {
	i.settings = setting
}

func (i *EC2Instance) Filter() error {
	if *i.State == ec2.InstanceStateNameTerminated {
		return fmt.Errorf("already terminated")
	}
	return nil
}

func (i *EC2Instance) Remove(_ context.Context) error {
	deleteTagsParams := &ec2.DeleteTagsInput{
		Resources: []*string{i.ID},
	}
	if _, err := i.svc.DeleteTags(deleteTagsParams); err != nil {
		return err
	}

	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{i.ID},
	}

	if _, err := i.svc.TerminateInstances(params); err != nil {
		var awsErr awserr.Error
		ok := errors.As(err, &awsErr)

		// Check for Termination Protection, disable it, and try termination again.
		if ok && awsErr.Code() == awsutil.ErrCodeOperationNotPermitted &&
			awsErr.Message() == "The instance '"+*i.ID+"' may not be "+
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
			awsErr.Message() == "The instance '"+*i.ID+"' may not be "+
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
		InstanceId: i.ID,
		DisableApiStop: &ec2.AttributeBooleanValue{
			Value: ptr.Bool(false),
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
		InstanceId: i.ID,
		DisableApiTermination: &ec2.AttributeBooleanValue{
			Value: ptr.Bool(false),
		},
	}
	_, err := i.svc.ModifyInstanceAttribute(params)
	if err != nil {
		return err
	}
	return nil
}

func (i *EC2Instance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(i)
}

func (i *EC2Instance) String() string {
	return *i.ID
}
