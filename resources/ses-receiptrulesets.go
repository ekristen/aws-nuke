package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ses" //nolint:staticcheck

	sdkerrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SESReceiptRuleSetResource = "SESReceiptRuleSet"

func init() {
	registry.Register(&registry.Registration{
		Name:     SESReceiptRuleSetResource,
		Scope:    nuke.Account,
		Resource: &SESReceiptRuleSet{},
		Lister:   &SESReceiptRuleSetLister{},
	})
}

type SESReceiptRuleSetLister struct{}

func (l *SESReceiptRuleSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ses.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ses.ListReceiptRuleSetsInput{}

	output, err := svc.ListReceiptRuleSets(params)
	if err != nil {
		// SES capabilities aren't the same in all regions, for example us-west-1 will throw InvalidAction
		// errors, but other regions work, this allows us to safely ignore these and yet log them in debug logs
		// should we need to troubleshoot.
		var awsError awserr.Error
		if errors.As(err, &awsError) {
			if awsError.Code() == awsutil.ErrCodeInvalidAction {
				return nil, sdkerrors.ErrSkipRequest(
					"Listing of SESReceiptFilter not supported in this region: " + *opts.Session.Config.Region)
			}
		}

		return nil, err
	}

	for _, ruleSet := range output.RuleSets {
		// Check active state
		ruleSetState := false
		ruleName := ruleSet.Name

		activeRuleSetOutput, err := svc.DescribeActiveReceiptRuleSet(&ses.DescribeActiveReceiptRuleSetInput{})
		if err != nil {
			return nil, err
		}
		if activeRuleSetOutput.Metadata == nil {
			ruleSetState = false
		} else if *ruleName == *activeRuleSetOutput.Metadata.Name {
			ruleSetState = true
		}

		resources = append(resources, &SESReceiptRuleSet{
			svc:           svc,
			name:          ruleName,
			activeRuleSet: ruleSetState,
		})
	}

	return resources, nil
}

type SESReceiptRuleSet struct {
	svc           *ses.SES
	name          *string
	activeRuleSet bool
}

func (f *SESReceiptRuleSet) Filter() error {
	if f.activeRuleSet {
		return fmt.Errorf("cannot delete active ruleset")
	}
	return nil
}

func (f *SESReceiptRuleSet) Remove(_ context.Context) error {
	_, err := f.svc.DeleteReceiptRuleSet(&ses.DeleteReceiptRuleSetInput{
		RuleSetName: f.name,
	})

	return err
}

func (f *SESReceiptRuleSet) String() string {
	return *f.name
}
