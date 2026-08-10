package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	quicksighttypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const QuickSightSubscriptionResource = "QuickSightSubscription"
const subscriptionNameWhenNotAvailable = "NOT_AVAILABLE"

func init() {
	registry.Register(&registry.Registration{
		Name:     QuickSightSubscriptionResource,
		Scope:    nuke.Account,
		Resource: &QuickSightSubscription{},
		Lister:   &QuickSightSubscriptionLister{},
		Settings: []string{
			"DisableTerminationProtection",
		},
	})
}

type QuickSightSubscriptionLister struct{}

type QuickSightSubscription struct {
	svc               *quicksight.Client
	cfg               aws.Config
	settings          *libsettings.Setting
	accountID         *string
	identityRegion    string
	name              *string
	notificationEmail *string
	edition           *string
	status            *string
}

func (l *QuickSightSubscriptionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := quicksight.NewFromConfig(*opts.Config)

	// DescribeAccountSubscription works from any region, so use it to check if
	// a subscription exists at all.
	describeOutput, err := svc.DescribeAccountSubscription(ctx, &quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: opts.AccountID,
	})
	if err != nil {
		if isResourceNotFound(err) {
			return resources, nil
		}
		return nil, err
	}

	subscriptionName := subscriptionNameWhenNotAvailable
	if describeOutput.AccountInfo.AccountName != nil {
		subscriptionName = *describeOutput.AccountInfo.AccountName
	}

	// Discover the identity region by calling DescribeAccountSettings.
	// Unlike DescribeAccountSubscription, this API enforces the identity region
	// and returns an error message containing the correct region if called from
	// the wrong endpoint. QuickSight does not provide a direct API to query the
	// identity region, so this is the only reliable discovery method.
	identityRegion := opts.Config.Region
	_, settingsErr := svc.DescribeAccountSettings(ctx, &quicksight.DescribeAccountSettingsInput{
		AwsAccountId: opts.AccountID,
	})
	if settingsErr != nil {
		if parsed := parseIdentityRegionFromError(settingsErr); parsed != "" {
			identityRegion = parsed
		}
	}

	// Create the client for the identity region
	identityCfg := opts.Config.Copy()
	identityCfg.Region = identityRegion
	identitySvc := quicksight.NewFromConfig(identityCfg)

	resources = append(resources, &QuickSightSubscription{
		svc:               identitySvc,
		cfg:               *opts.Config,
		accountID:         opts.AccountID,
		identityRegion:    identityRegion,
		name:              &subscriptionName,
		notificationEmail: describeOutput.AccountInfo.NotificationEmail,
		edition:           editionToString(describeOutput.AccountInfo.Edition),
		status:            describeOutput.AccountInfo.AccountSubscriptionStatus,
	})

	return resources, nil
}

func (r *QuickSightSubscription) Remove(ctx context.Context) error {
	// r.svc is already configured for the identity region (set during List)

	// Disable termination protection before attempting deletion
	if err := r.disableTerminationProtection(ctx); err != nil {
		// If we still get an identity region error, try the parsed region
		if parsed := parseIdentityRegionFromError(err); parsed != "" {
			regionalCfg := r.cfg.Copy()
			regionalCfg.Region = parsed
			r.svc = quicksight.NewFromConfig(regionalCfg)
			r.identityRegion = parsed
			_ = r.disableTerminationProtection(ctx)
		}
	}

	// Delete the subscription
	_, err := r.svc.DeleteAccountSubscription(ctx, &quicksight.DeleteAccountSubscriptionInput{
		AwsAccountId: r.accountID,
	})
	if err != nil {
		// Handle identity region redirect on delete
		if parsed := parseIdentityRegionFromError(err); parsed != "" {
			regionalCfg := r.cfg.Copy()
			regionalCfg.Region = parsed
			regionalSvc := quicksight.NewFromConfig(regionalCfg)
			_ = r.disableTerminationProtectionWithSvc(ctx, regionalSvc)
			_, err = regionalSvc.DeleteAccountSubscription(ctx, &quicksight.DeleteAccountSubscriptionInput{
				AwsAccountId: r.accountID,
			})
			return err
		}
		return err
	}

	return nil
}

func (r *QuickSightSubscription) disableTerminationProtection(ctx context.Context) error {
	return r.disableTerminationProtectionWithSvc(ctx, r.svc)
}

func (r *QuickSightSubscription) disableTerminationProtectionWithSvc(ctx context.Context, svc *quicksight.Client) error {
	describeSettingsOutput, err := svc.DescribeAccountSettings(ctx, &quicksight.DescribeAccountSettingsInput{
		AwsAccountId: r.accountID,
	})
	if err != nil {
		return err
	}

	if describeSettingsOutput.AccountSettings.TerminationProtectionEnabled {
		_, err = svc.UpdateAccountSettings(ctx, &quicksight.UpdateAccountSettingsInput{
			AwsAccountId:                 r.accountID,
			DefaultNamespace:             describeSettingsOutput.AccountSettings.DefaultNamespace,
			NotificationEmail:            describeSettingsOutput.AccountSettings.NotificationEmail,
			TerminationProtectionEnabled: false,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// parseIdentityRegionFromError extracts the identity region from a QuickSight error message.
// QuickSight returns errors like: "Operation is being called from endpoint us-east-1,
// but your identity region is us-west-2. Please use the us-west-2 endpoint."
// This is the only way to discover the identity region, as no QuickSight API
// returns it directly.
func parseIdentityRegionFromError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	const marker = "identity region is "
	idx := strings.Index(msg, marker)
	if idx == -1 {
		return ""
	}
	regionStart := idx + len(marker)
	var region strings.Builder
	for i := regionStart; i < len(msg); i++ {
		c := msg[i]
		if c == '.' || c == ' ' || c == '"' {
			break
		}
		region.WriteByte(c)
	}
	result := region.String()
	if result != "" && strings.Contains(result, "-") {
		return result
	}
	return ""
}

func isResourceNotFound(err error) bool {
	return strings.Contains(err.Error(), "ResourceNotFoundException")
}

func editionToString(edition quicksighttypes.Edition) *string {
	s := string(edition)
	return &s
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

	if *r.name == subscriptionNameWhenNotAvailable {
		return fmt.Errorf("subscription name is not available yet")
	}
	return nil
}

func (r *QuickSightSubscription) Settings(setting *libsettings.Setting) {
	r.settings = setting
}
