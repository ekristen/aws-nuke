package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/accessanalyzer"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AccessAnalyzerArchiveRuleResource = "AccessAnalyzerArchiveRule"

func init() {
	registry.Register(&registry.Registration{
		Name:     AccessAnalyzerArchiveRuleResource,
		Scope:    nuke.Account,
		Resource: &ArchiveRule{},
		Lister:   &AccessAnalyzerArchiveRuleLister{},
		DeprecatedAliases: []string{
			"ArchiveRule",
		},
	})
}

type ArchiveRule struct {
	svc          *accessanalyzer.AccessAnalyzer
	RuleName     *string `description:"The name of the archive rule"`
	AnalyzerName *string `description:"The name of the analyzer the rule is associated with"`
}

func (r *ArchiveRule) Remove(_ context.Context) error {
	_, err := r.svc.DeleteArchiveRule(&accessanalyzer.DeleteArchiveRuleInput{
		AnalyzerName: r.AnalyzerName,
		RuleName:     r.RuleName,
	})

	return err
}

func (r *ArchiveRule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ArchiveRule) String() string {
	return *r.RuleName
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
			AnalyzerName: a.Name,
		}

		err = svc.ListArchiveRulesPages(params,
			func(page *accessanalyzer.ListArchiveRulesOutput, lastPage bool) bool {
				for _, archiveRule := range page.ArchiveRules {
					resources = append(resources, &ArchiveRule{
						svc:          svc,
						RuleName:     archiveRule.RuleName,
						AnalyzerName: a.Name,
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
