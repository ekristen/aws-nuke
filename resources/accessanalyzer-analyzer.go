package resources

import (
	"context"
	"errors"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/accessanalyzer" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AccessAnalyzerResource = "AccessAnalyzer"

// securityHubV2AnalyzerPrefix is the name prefix of the unused access analyzer that Security Hub V2 provisions on the
// account's behalf. It is owned by the service and cannot be deleted, not even once Security Hub V2 is disabled.
const securityHubV2AnalyzerPrefix = "_AccessAnalyzerForSecurityHubV2"

func init() {
	registry.Register(&registry.Registration{
		Name:                AccessAnalyzerResource,
		Scope:               nuke.Account,
		Resource:            &AccessAnalyzer{},
		Lister:              &AccessAnalyzerLister{},
		AlternativeResource: "AWS::AccessAnalyzer::Analyzer",
	})
}

type AccessAnalyzerLister struct{}

func (l *AccessAnalyzerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := accessanalyzer.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &accessanalyzer.ListAnalyzersInput{}

	if err := svc.ListAnalyzersPages(params,
		func(page *accessanalyzer.ListAnalyzersOutput, lastPage bool) bool {
			for _, analyzer := range page.Analyzers {
				resources = append(resources, &AccessAnalyzer{
					svc:    svc,
					ARN:    analyzer.Arn,
					Name:   analyzer.Name,
					Status: analyzer.Status,
					Type:   analyzer.Type,
					Tags:   analyzer.Tags,
				})
			}
			return true
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

type AccessAnalyzer struct {
	svc    *accessanalyzer.AccessAnalyzer
	ARN    *string            `description:"The ARN of the analyzer"`
	Name   *string            `description:"The name of the analyzer"`
	Status *string            `description:"The status of the analyzer"`
	Type   *string            `description:"The type of the analyzer"`
	Tags   map[string]*string `description:"The tags of the analyzer"`
}

func (r *AccessAnalyzer) Filter() error {
	if strings.Contains(ptr.ToString(r.Name), "ORGANIZATION") {
		return errors.New("cannot delete organization analyzer")
	}
	if strings.HasPrefix(ptr.ToString(r.Name), securityHubV2AnalyzerPrefix) {
		return errors.New("cannot delete analyzer managed by security hub v2")
	}
	return nil
}

func (r *AccessAnalyzer) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAnalyzer(&accessanalyzer.DeleteAnalyzerInput{AnalyzerName: r.Name})

	return err
}

func (r *AccessAnalyzer) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AccessAnalyzer) String() string {
	return *r.Name
}
