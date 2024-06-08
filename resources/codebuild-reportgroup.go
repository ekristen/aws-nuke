package resources

import (
	"context"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeBuildReportGroupResource = "CodeBuildReportGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   CodeBuildReportGroupResource,
		Scope:  nuke.Account,
		Lister: &CodebuildReportGroupLister{},
		DependsOn: []string{
			CodeBuildReportResource,
		},
	})
}

type CodebuildReportGroupLister struct{}

func (l *CodebuildReportGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := codebuild.New(opts.Session)
	var resources []resource.Resource

	res, err := svc.ListReportGroups(&codebuild.ListReportGroupsInput{})
	if err != nil {
		return nil, err
	}

	for _, arn := range res.ReportGroups {
		resources = append(resources, &CodebuildReportGroup{
			svc: svc,
			arn: arn,
		})
	}

	return resources, nil
}

type CodebuildReportGroup struct {
	svc *codebuild.CodeBuild
	arn *string
}

func (r *CodebuildReportGroup) Name() string {
	return strings.Split(*r.arn, "report-group/")[1]
}

func (r *CodebuildReportGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteReportGroup(&codebuild.DeleteReportGroupInput{
		Arn:           r.arn,
		DeleteReports: ptr.Bool(true),
	})
	return err
}

func (r *CodebuildReportGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.Name())
	return properties
}

func (r *CodebuildReportGroup) String() string {
	return r.Name()
}
