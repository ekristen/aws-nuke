package awsutil

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"           //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/endpoints" //nolint:staticcheck
	"github.com/gotidy/ptr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/sts" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/pkg/config"
)

type Account struct {
	*Credentials

	id              string
	arn             string
	userID          string
	aliases         []string
	regions         []string
	disabledRegions []string
}

func NewAccount(creds *Credentials, customEndpoints config.CustomEndpoints) (*Account, error) {
	creds.CustomEndpoints = customEndpoints
	account := Account{
		Credentials: creds,
	}

	customStackSupportSTSAndIAM := true
	if customEndpoints.GetRegion(DefaultRegionID) != nil {
		if customEndpoints.GetURL(DefaultRegionID, "sts") == "" {
			customStackSupportSTSAndIAM = false
		} else if customEndpoints.GetURL(DefaultRegionID, "iam") == "" {
			customStackSupportSTSAndIAM = false
		}
	}
	if !customStackSupportSTSAndIAM {
		account.id = "account-id-of-custom-region-" + DefaultRegionID
		account.aliases = []string{account.id}
		return &account, nil
	}

	defaultSession, err := account.NewSession(DefaultRegionID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create default session in %s", DefaultRegionID)
	}

	identityOutput, err := sts.New(defaultSession, &aws.Config{STSRegionalEndpoint: endpoints.RegionalSTSEndpoint}).GetCallerIdentity(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed get caller identity")
	}

	regionsOutput, err := ec2.New(defaultSession).DescribeRegions(&ec2.DescribeRegionsInput{
		AllRegions: ptr.Bool(true),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get regions")
	}

	globalSession, err := account.NewSession(GlobalRegionID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create global session in %s", GlobalRegionID)
	}

	aliasesOutput, err := iam.New(globalSession).ListAccountAliases(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed get account alias")
	}

	var aliases []string
	for _, alias := range aliasesOutput.AccountAliases {
		if alias != nil {
			aliases = append(aliases, *alias)
		}
	}

	regions := []string{"global"}
	var disabledRegions []string
	for _, region := range regionsOutput.Regions {
		logrus.Debugf("region: %s, status: %s",
			ptr.ToString(region.RegionName), ptr.ToString(region.OptInStatus))

		if ptr.ToString(region.OptInStatus) == "not-opted-in" {
			disabledRegions = append(disabledRegions, *region.RegionName)
		} else {
			regions = append(regions, *region.RegionName)
		}
	}

	account.id = ptr.ToString(identityOutput.Account)
	account.arn = ptr.ToString(identityOutput.Arn)
	account.userID = ptr.ToString(identityOutput.UserId)
	account.aliases = aliases
	account.regions = regions
	account.disabledRegions = disabledRegions

	return &account, nil
}

// ID returns the account ID
func (a *Account) ID() string {
	return a.id
}

// ARN returns the STS Authenticated ARN for the account
func (a *Account) ARN() string {
	return a.arn
}

// UserID returns the authenticated user ID
func (a *Account) UserID() string {
	return a.userID
}

// Alias returns the first alias for the account
func (a *Account) Alias() string {
	if len(a.aliases) == 0 {
		return fmt.Sprintf("no-alias-%s", a.ID())
	}

	return a.aliases[0]
}

// Aliases returns the list of aliases for the account
func (a *Account) Aliases() []string {
	return a.aliases
}

func (a *Account) ResourceTypeToServiceType(regionName, resourceType string) string {
	customRegion := a.CustomEndpoints.GetRegion(regionName)
	if customRegion == nil {
		return "-" // standard public AWS.
	}
	for _, e := range customRegion.Services {
		if strings.HasPrefix(strings.ToLower(resourceType), e.Service) {
			return e.Service
		}
	}
	return ""
}

// Regions returns the list of regions that are enabled for the account
func (a *Account) Regions() []string {
	return a.regions
}

// DisabledRegions returns the list of regions that are disabled for the account
func (a *Account) DisabledRegions() []string {
	return a.disabledRegions
}
