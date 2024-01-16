package nuke

import (
	"fmt"
	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/featureflag"
	"github.com/ekristen/libnuke/pkg/filter"
	sdknuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/utils"
	"github.com/sirupsen/logrus"
	"time"
)

type Parameters struct {
	sdknuke.Parameters

	Targets      []string
	Excludes     []string
	CloudControl []string
}

type Nuke struct {
	*sdknuke.Nuke
	Parameters Parameters
	Config     *config.Nuke
	Account    awsutil.Account
}

func (n *Nuke) Prompt() error {
	forceSleep := time.Duration(n.Parameters.ForceSleep) * time.Second

	fmt.Printf("Do you really want to nuke the account with "+
		"the ID %s and the alias '%s'?\n", n.Account.ID(), n.Account.Alias())
	if n.Parameters.Force {
		fmt.Printf("Waiting %v before continuing.\n", forceSleep)
		time.Sleep(forceSleep)
	} else {
		fmt.Printf("Do you want to continue? Enter account alias to continue.\n")
		if err := utils.Prompt(n.Account.Alias()); err != nil {
			return err
		}
	}

	return nil
}

func New(params Parameters, config *config.Nuke, filters filter.Filters, account awsutil.Account) *Nuke {
	n := Nuke{
		Nuke:       sdknuke.New(params.Parameters, filters),
		Parameters: params,
		Config:     config,
		Account:    account,
	}

	n.SetLogger(logrus.WithField("component", "nuke"))

	defaultValue := featureflag.Bool(false)

	n.RegisterFeatureFlags("DisableEC2InstanceStopProtection", defaultValue, featureflag.Bool(config.FeatureFlags.DisableEC2InstanceStopProtection))
	n.RegisterFeatureFlags("ForceDeleteLightsailAddOns", defaultValue, featureflag.Bool(config.FeatureFlags.ForceDeleteLightsailAddOns))
	n.RegisterFeatureFlags("DisableDeletionProtection_RDSInstance", defaultValue, featureflag.Bool(config.FeatureFlags.DisableDeletionProtection.RDSInstance))
	n.RegisterFeatureFlags("DisableDeletionProtection_EC2Instance", defaultValue, featureflag.Bool(config.FeatureFlags.DisableDeletionProtection.EC2Instance))
	n.RegisterFeatureFlags("DisableDeletionProtection_ELBv2", defaultValue, featureflag.Bool(config.FeatureFlags.DisableDeletionProtection.ELBv2))
	n.RegisterFeatureFlags("DisableDeletionProtection_CloudformationStack", defaultValue, featureflag.Bool(config.FeatureFlags.DisableDeletionProtection.CloudformationStack))
	n.RegisterFeatureFlags("DisableDeletionProtection_QLDBLedger", defaultValue, featureflag.Bool(config.FeatureFlags.DisableDeletionProtection.QLDBLedger))

	return &n
}
