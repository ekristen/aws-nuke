package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/quicksight/quicksightiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const quickSightSubscriptionResource = "QuicksightSubscription"
const subscriptionNameWhenNotAvailable = "NOT_AVAILABLE"

func init() {
	registry.Register(&registry.Registration{
		Name:   quickSightSubscriptionResource,
		Scope:  nuke.Account,
		Lister: &QuickSightSubscriptionLister{},
	})
}

type QuickSightSubscriptionLister struct{
	stsService stsiface.STSAPI
	quicksightService quicksightiface.QuickSightAPI
}

type QuicksightSubscription struct {
	svc               quicksightiface.QuickSightAPI
	accountId         *string
	name  			  *string
	notificationEmail *string
	edition           *string
	status            *string
}

func (listener *QuickSightSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	
	var stsSvc stsiface.STSAPI
	if listener.stsService != nil {
		stsSvc = listener.stsService
	} else {
		stsSvc = sts.New(opts.Session)
	}

	var quicksightSvc quicksightiface.QuickSightAPI
	if listener.quicksightService != nil {
		quicksightSvc = listener.quicksightService
	} else {
		quicksightSvc = quicksight.New(opts.Session)
	}
	
	callerID, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	accountId := callerID.Account

	var resources []resource.Resource

	describeSubscriptionOutput, err := quicksightSvc.DescribeAccountSubscription(&quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: accountId,
	})

	if err != nil {
		var resoureceNotFoundException *quicksight.ResourceNotFoundException
		if !errors.As(err, &resoureceNotFoundException) {
			return nil, err
		}
		return resources, nil
	}

	//The account name is only available some time later after the Subscription creation.
	subscriptionName := subscriptionNameWhenNotAvailable
	if describeSubscriptionOutput.AccountInfo.AccountName != nil {
		subscriptionName = *describeSubscriptionOutput.AccountInfo.AccountName
	}

	resources = append(resources, &QuicksightSubscription{
		svc:               quicksightSvc,
		accountId:         accountId,
		name:			   &subscriptionName,
		notificationEmail: describeSubscriptionOutput.AccountInfo.NotificationEmail,
		edition:           describeSubscriptionOutput.AccountInfo.Edition,
		status:            describeSubscriptionOutput.AccountInfo.AccountSubscriptionStatus,
	})

	return resources, nil
}

func (subscription *QuicksightSubscription) Remove(_ context.Context) error {
	terminateProtectionEnabled := false

	describeSettingsOutput, err := subscription.svc.DescribeAccountSettings(&quicksight.DescribeAccountSettingsInput{
		AwsAccountId: subscription.accountId,
	})
	if err != nil {
		return err
	}

	if *describeSettingsOutput.AccountSettings.TerminationProtectionEnabled {
		updateSettingsInput := quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 subscription.accountId,
			DefaultNamespace:             describeSettingsOutput.AccountSettings.DefaultNamespace,
			NotificationEmail:            describeSettingsOutput.AccountSettings.NotificationEmail,
			TerminationProtectionEnabled: &terminateProtectionEnabled,
		}

		_, err = subscription.svc.UpdateAccountSettings(&updateSettingsInput)
		if err != nil {
			return err
		}
	}

	_, err = subscription.svc.DeleteAccountSubscription(&quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: subscription.accountId,
	})
	if err != nil {
		return err
	}

	return nil
}

func (subscription *QuicksightSubscription) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Edition", subscription.edition).
		Set("NotificationEmail", subscription.notificationEmail).
		Set("Name", subscription.name).
		Set("Status", subscription.status)

	return properties
}

func (subscription *QuicksightSubscription) String() string {
	return *subscription.name
}

func (subscription *QuicksightSubscription) Filter() error {
	if *subscription.status != "ACCOUNT_CREATED" {
		return fmt.Errorf("subscription is not active")
	}

	//Since the subscription name is an important value to identify the resource, it will wait till it is available
	if *subscription.name == subscriptionNameWhenNotAvailable {
		return fmt.Errorf("subscription name is not available yet")
	}
	return nil
}
