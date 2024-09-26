package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const PinpointPhoneNumberResource = "PinpointPhoneNumber"

func init() {
	registry.Register(&registry.Registration{
		Name:   PinpointPhoneNumberResource,
		Scope:  nuke.Account,
		Lister: &PinpointPhoneNumberLister{},
	})
}

type PinpointPhoneNumberLister struct{}

func (l *PinpointPhoneNumberLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := pinpointsmsvoicev2.New(opts.Session)

	resp, err := svc.DescribePhoneNumbers(&pinpointsmsvoicev2.DescribePhoneNumbersInput{})
	if err != nil {
		return nil, err
	}

	numbers := make([]resource.Resource, 0)
	for _, number := range resp.PhoneNumbers {
		numbers = append(numbers, &PinpointPhoneNumber{
			svc: svc,
			ID:  number.PhoneNumberId,
		})
	}

	return numbers, nil
}

type PinpointPhoneNumber struct {
	svc *pinpointsmsvoicev2.PinpointSMSVoiceV2
	ID  *string
}

func (r *PinpointPhoneNumber) Remove(_ context.Context) error {
	params := &pinpointsmsvoicev2.ReleasePhoneNumberInput{
		PhoneNumberId: r.ID,
	}

	_, err := r.svc.ReleasePhoneNumber(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *PinpointPhoneNumber) String() string {
	return *r.ID
}
