package resources

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/scheduler" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SchedulerScheduleResource = "SchedulerSchedule"

func init() {
	registry.Register(&registry.Registration{
		Name:     SchedulerScheduleResource,
		Scope:    nuke.Account,
		Resource: &SchedulerSchedule{},
		Lister:   &SchedulerScheduleLister{},
	})
}

type SchedulerScheduleLister struct{}

func (l *SchedulerScheduleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := scheduler.New(opts.Session)
	var resources []resource.Resource

	res, err := svc.ListSchedules(&scheduler.ListSchedulesInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.Schedules {
		resources = append(resources, &SchedulerSchedule{
			svc:          svc,
			Name:         p.Name,
			State:        p.State,
			GroupName:    p.GroupName,
			CreationDate: p.CreationDate,
			ModifiedDate: p.LastModificationDate,
		})
	}

	return resources, nil
}

type SchedulerSchedule struct {
	svc          *scheduler.Scheduler
	Name         *string
	State        *string
	GroupName    *string
	CreationDate *time.Time
	ModifiedDate *time.Time
}

func (r *SchedulerSchedule) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSchedule(&scheduler.DeleteScheduleInput{
		Name:        r.Name,
		GroupName:   r.GroupName,
		ClientToken: ptr.String(uuid.Must(uuid.NewUUID()).String()),
	})
	return err
}

func (r *SchedulerSchedule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SchedulerSchedule) String() string {
	return *r.Name
}
