package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMUserMFADeviceResource = "IAMUserMFADevice"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMUserMFADeviceResource,
		Scope:    nuke.Account,
		Resource: &IAMUserMFADevice{},
		Lister:   &IAMUserMFADeviceLister{},
	})
}

type IAMUserMFADeviceLister struct {
	mockSvc iamiface.IAMAPI
}

func (l *IAMUserMFADeviceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	var svc iamiface.IAMAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = iam.New(opts.Session)
	}

	allUsers, err := ListIAMUsers(svc)
	if err != nil {
		return nil, err
	}

	for _, user := range allUsers {
		res, listErr := svc.ListMFADevices(&iam.ListMFADevicesInput{
			UserName: user.UserName,
		})
		if listErr != nil {
			return nil, listErr
		}

		for _, p := range res.MFADevices {
			nameParts := strings.Split(*p.SerialNumber, "/")
			name := nameParts[len(nameParts)-1]

			resources = append(resources, &IAMUserMFADevice{
				svc:          svc,
				Name:         ptr.String(name),
				UserName:     p.UserName,
				SerialNumber: p.SerialNumber,
				EnableDate:   p.EnableDate,
			})
		}
	}

	return resources, nil
}

type IAMUserMFADevice struct {
	svc          iamiface.IAMAPI
	Name         *string
	UserName     *string
	SerialNumber *string
	EnableDate   *time.Time
}

func (r *IAMUserMFADevice) Remove(_ context.Context) error {
	_, err := r.svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
		UserName:     r.UserName,
		SerialNumber: r.SerialNumber,
	})
	return err
}

func (r *IAMUserMFADevice) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IAMUserMFADevice) String() string {
	return fmt.Sprintf("%s -> %s", *r.UserName, *r.Name)
}
