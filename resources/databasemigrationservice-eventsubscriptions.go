package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DatabaseMigrationServiceEventSubscriptionResource = "DatabaseMigrationServiceEventSubscription"

func init() {
	resource.Register(resource.Registration{
		Name:   DatabaseMigrationServiceEventSubscriptionResource,
		Scope:  nuke.Account,
		Lister: &DatabaseMigrationServiceEventSubscriptionLister{},
	})
}

type DatabaseMigrationServiceEventSubscriptionLister struct{}

func (l *DatabaseMigrationServiceEventSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := databasemigrationservice.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &databasemigrationservice.DescribeEventSubscriptionsInput{
		MaxRecords: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeEventSubscriptions(params)
		if err != nil {
			return nil, err
		}

		for _, eventSubscription := range output.EventSubscriptionsList {
			resources = append(resources, &DatabaseMigrationServiceEventSubscription{
				svc:              svc,
				subscriptionName: eventSubscription.CustSubscriptionId,
			})
		}

		if output.Marker == nil {
			break
		}

		params.Marker = output.Marker
	}

	return resources, nil
}

type DatabaseMigrationServiceEventSubscription struct {
	svc              *databasemigrationservice.DatabaseMigrationService
	subscriptionName *string
}

func (f *DatabaseMigrationServiceEventSubscription) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEventSubscription(&databasemigrationservice.DeleteEventSubscriptionInput{
		SubscriptionName: f.subscriptionName,
	})

	return err
}

func (f *DatabaseMigrationServiceEventSubscription) String() string {
	return *f.subscriptionName
}
