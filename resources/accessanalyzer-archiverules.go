package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/accessanalyzer"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AccessAnalyzerArchiveRuleResource = "ArchiveRule"

func init() {
	resource.Register(&resource.Registration{
		Name:   AccessAnalyzerArchiveRuleResource,
		Scope:  nuke.Account,
		Lister: &AccessAnalyzerArchiveRuleLister{},
	})
}

type ArchiveRule struct {
	svc          *accessanalyzer.AccessAnalyzer
	ruleName     string
	analyzerName string
}

func (a *ArchiveRule) Remove(_ context.Context) error {
	_, err := a.svc.DeleteArchiveRule(&accessanalyzer.DeleteArchiveRuleInput{
		AnalyzerName: &a.analyzerName,
		RuleName:     &a.ruleName,
	})

	return err
}

func (a *ArchiveRule) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("RuleName", a.ruleName)
	properties.Set("AnalyzerName", a.analyzerName)

	return properties
}

func (a *ArchiveRule) String() string {
	return a.ruleName
}

// ---------------------------

type AccessAnalyzerArchiveRuleLister struct{}

func (l *AccessAnalyzerArchiveRuleLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := accessanalyzer.New(opts.Session)

	lister := &AccessAnalyzerLister{}
	analyzers, err := lister.List(ctx, o)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	for _, analyzer := range analyzers {
		a, ok := analyzer.(*AccessAnalyzer)
		if !ok {
			continue
		}

		params := &accessanalyzer.ListArchiveRulesInput{
			AnalyzerName: &a.name,
		}

		err = svc.ListArchiveRulesPages(params,
			func(page *accessanalyzer.ListArchiveRulesOutput, lastPage bool) bool {
				for _, archiveRule := range page.ArchiveRules {
					resources = append(resources, &ArchiveRule{
						svc:          svc,
						ruleName:     *archiveRule.RuleName,
						analyzerName: a.name,
					})
				}
				return true
			})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}
