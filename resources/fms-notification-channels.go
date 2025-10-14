package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/fms" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FMSNotificationChannelResource = "FMSNotificationChannel"

func init() {
	registry.Register(&registry.Registration{
		Name:     FMSNotificationChannelResource,
		Scope:    nuke.Account,
		Resource: &FMSNotificationChannel{},
		Lister:   &FMSNotificationChannelLister{},
	})
}

type FMSNotificationChannelLister struct{}

func (l *FMSNotificationChannelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := fms.New(opts.Session)
	resources := make([]resource.Resource, 0)

	if _, err := svc.GetNotificationChannel(&fms.GetNotificationChannelInput{}); err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			if aerr.Code() != fms.ErrCodeResourceNotFoundException {
				return nil, err
			}
		}
	} else {
		resources = append(resources, &FMSNotificationChannel{
			svc: svc,
		})
	}

	return resources, nil
}

type FMSNotificationChannel struct {
	svc *fms.FMS
}

func (f *FMSNotificationChannel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteNotificationChannel(&fms.DeleteNotificationChannelInput{})

	return err
}

func (f *FMSNotificationChannel) String() string {
	return "fms-notification-channel"
}

func (f *FMSNotificationChannel) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("NotificationChannelEnabled", "true")
	return properties
}
