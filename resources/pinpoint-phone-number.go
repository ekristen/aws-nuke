package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const PinpointPhoneNumberResource = "PinpointPhoneNumber"

func init() {
	registry.Register(&registry.Registration{
		Name:     PinpointPhoneNumberResource,
		Scope:    nuke.Account,
		Resource: &PinpointPhoneNumber{},
		Lister:   &PinpointPhoneNumberLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
	})
}

type PinpointPhoneNumberLister struct{}

func (l *PinpointPhoneNumberLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := pinpointsmsvoicev2.New(opts.Session)

	params := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}

	for {
		resp, err := svc.DescribePhoneNumbers(params)
		if err != nil {
			return nil, err
		}

		for _, number := range resp.PhoneNumbers {
			resources = append(resources, &PinpointPhoneNumber{
				svc:         svc,
				settings:    &settings.Setting{},
				ID:          number.PhoneNumberId,
				Status:      number.Status,
				CreatedDate: number.CreatedTimestamp,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type PinpointPhoneNumber struct {
	svc                       *pinpointsmsvoicev2.PinpointSMSVoiceV2
	settings                  *settings.Setting
	ID                        *string
	Status                    *string
	CreatedDate               *time.Time
	deletionProtectionEnabled *bool
}

func (r *PinpointPhoneNumber) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PinpointPhoneNumber) Remove(_ context.Context) error {
	if r.settings.GetBool("DisableDeletionProtection") && ptr.ToBool(r.deletionProtectionEnabled) {
		_, err := r.svc.UpdatePhoneNumber(&pinpointsmsvoicev2.UpdatePhoneNumberInput{
			PhoneNumberId:             r.ID,
			DeletionProtectionEnabled: ptr.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.ReleasePhoneNumber(&pinpointsmsvoicev2.ReleasePhoneNumberInput{
		PhoneNumberId: r.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *PinpointPhoneNumber) Settings(setting *settings.Setting) {
	r.settings = setting
}

func (r *PinpointPhoneNumber) String() string {
	return *r.ID
}
