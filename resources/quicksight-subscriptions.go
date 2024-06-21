package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/quicksight/quicksightiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"
)

const QuickSightSubscriptionResource = "QuickSightSubscription"
const subscriptionNameWhenNotAvailable = "NOT_AVAILABLE"

func init() {
	registry.Register(&registry.Registration{
		Name:   QuickSightSubscriptionResource,
		Scope:  nuke.Account,
		Lister: &QuickSightSubscriptionLister{},
	})
}

type QuickSightSubscriptionLister struct {
	stsService        stsiface.STSAPI
	quicksightService quicksightiface.QuickSightAPI
}

type QuickSightSubscription struct {
	svc               quicksightiface.QuickSightAPI
	accountID         *string
	name              *string
	notificationEmail *string
	edition           *string
	status            *string
	settings          *libsettings.Setting
}

func (l *QuickSightSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var stsSvc stsiface.STSAPI
	if l.stsService != nil {
		stsSvc = l.stsService
	} else {
		stsSvc = sts.New(opts.Session)
	}

	var quicksightSvc quicksightiface.QuickSightAPI
	if l.quicksightService != nil {
		quicksightSvc = l.quicksightService
	} else {
		quicksightSvc = quicksight.New(opts.Session)
	}

	callerID, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	accountID := callerID.Account

	var resources []resource.Resource

	describeSubscriptionOutput, err := quicksightSvc.DescribeAccountSubscription(&quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: accountID,
	})

	if err != nil {
		var resoureceNotFoundException *quicksight.ResourceNotFoundException
		if !errors.As(err, &resoureceNotFoundException) {
			return nil, err
		}
		return resources, nil
	}

	// The account name is only available some time later after the Subscription creation.
	subscriptionName := subscriptionNameWhenNotAvailable
	if describeSubscriptionOutput.AccountInfo.AccountName != nil {
		subscriptionName = *describeSubscriptionOutput.AccountInfo.AccountName
	}

	resources = append(resources, &QuickSightSubscription{
		svc:               quicksightSvc,
		accountID:         accountID,
		name:              &subscriptionName,
		notificationEmail: describeSubscriptionOutput.AccountInfo.NotificationEmail,
		edition:           describeSubscriptionOutput.AccountInfo.Edition,
		status:            describeSubscriptionOutput.AccountInfo.AccountSubscriptionStatus,
	})

	return resources, nil
}

func (r *QuickSightSubscription) Remove(_ context.Context) error {
	if r.settings != nil && r.settings.GetBool("DisableTerminationProtection") {
		err := r.DisableTerminationProtection()
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: r.accountID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *QuickSightSubscription) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Edition", r.edition).
		Set("NotificationEmail", r.notificationEmail).
		Set("Name", r.name).
		Set("Status", r.status)

	return properties
}

func (r *QuickSightSubscription) String() string {
	return *r.name
}

func (r *QuickSightSubscription) Filter() error {
	if *r.status != "ACCOUNT_CREATED" {
		return fmt.Errorf("subscription is not active")
	}

	// Since the subscription name is an important value to identify the resource, it will wait till it is available
	if *r.name == subscriptionNameWhenNotAvailable {
		return fmt.Errorf("subscription name is not available yet")
	}
	return nil
}

func (r *QuickSightSubscription) Settings(setting *libsettings.Setting) {
	r.settings = setting
}

func (r *QuickSightSubscription) DisableTerminationProtection() error {
	terminateProtectionEnabled := false
	describeSettingsOutput, err := r.svc.DescribeAccountSettings(&quicksight.DescribeAccountSettingsInput{
		AwsAccountId: r.accountID,
	})
	if err != nil {
		return err
	}

	if *describeSettingsOutput.AccountSettings.TerminationProtectionEnabled {
		updateSettingsInput := quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 r.accountID,
			DefaultNamespace:             describeSettingsOutput.AccountSettings.DefaultNamespace,
			NotificationEmail:            describeSettingsOutput.AccountSettings.NotificationEmail,
			TerminationProtectionEnabled: &terminateProtectionEnabled,
		}

		_, err = r.svc.UpdateAccountSettings(&updateSettingsInput)
		if err != nil {
			return err
		}
	}
	return nil
}
