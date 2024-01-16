package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const AccessAnalyzerResource = "AccessAnalyzer"

type AccessAnalyzer struct {
	svc    *accessanalyzer.AccessAnalyzer
	arn    string
	name   string
	status string
	tags   map[string]*string
}

func init() {
	resource.Register(resource.Registration{
		Name:   AccessAnalyzerResource,
		Scope:  nuke.Account,
		Lister: &AccessAnalyzerLister{},
	}, nuke.MapCloudControl("AWS::AccessAnalyzer::Analyzer"))
}

type AccessAnalyzerLister struct{}

func (l *AccessAnalyzerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := accessanalyzer.New(opts.Session)

	params := &accessanalyzer.ListAnalyzersInput{
		Type: aws.String("ACCOUNT"),
	}

	resources := make([]resource.Resource, 0)
	if err := svc.ListAnalyzersPages(params,
		func(page *accessanalyzer.ListAnalyzersOutput, lastPage bool) bool {
			for _, analyzer := range page.Analyzers {
				resources = append(resources, &AccessAnalyzer{
					svc:    svc,
					arn:    *analyzer.Arn,
					name:   *analyzer.Name,
					status: *analyzer.Status,
					tags:   analyzer.Tags,
				})
			}
			return true
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

func (a *AccessAnalyzer) Remove(_ context.Context) error {
	_, err := a.svc.DeleteAnalyzer(&accessanalyzer.DeleteAnalyzerInput{AnalyzerName: &a.name})

	return err
}

func (a *AccessAnalyzer) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ARN", a.arn)
	properties.Set("Name", a.name)
	properties.Set("Status", a.status)
	for k, v := range a.tags {
		properties.SetTag(&k, v)
	}

	return properties
}

func (a *AccessAnalyzer) String() string {
	return a.name
}
