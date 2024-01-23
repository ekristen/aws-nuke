package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/guardduty"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GuardDutyDetectorResource = "GuardDutyDetector"

func init() {
	resource.Register(&resource.Registration{
		Name:   GuardDutyDetectorResource,
		Scope:  nuke.Account,
		Lister: &GuardDutyDetectorLister{},
	})
}

type GuardDutyDetectorLister struct{}

func (l *GuardDutyDetectorLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := guardduty.New(opts.Session)

	detectors := make([]resource.Resource, 0)

	params := &guardduty.ListDetectorsInput{}

	err := svc.ListDetectorsPages(params, func(page *guardduty.ListDetectorsOutput, lastPage bool) bool {
		for _, out := range page.DetectorIds {
			detectors = append(detectors, &GuardDutyDetector{
				svc: svc,
				id:  out,
			})
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return detectors, nil
}

type GuardDutyDetector struct {
	svc *guardduty.GuardDuty
	id  *string
}

func (r *GuardDutyDetector) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDetector(&guardduty.DeleteDetectorInput{
		DetectorId: r.id,
	})
	return err
}

func (r *GuardDutyDetector) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("DetectorID", r.id)
	return properties
}

func (r *GuardDutyDetector) String() string {
	return *r.id
}
