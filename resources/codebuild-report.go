package resources

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/service/codebuild" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CodeBuildReportResource = "CodeBuildReport"

func init() {
	registry.Register(&registry.Registration{
		Name:     CodeBuildReportResource,
		Scope:    nuke.Account,
		Resource: &CodeBuildReport{},
		Lister:   &CodeBuildReportLister{},
	})
}

type CodeBuildReportLister struct{}

func (l *CodeBuildReportLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := codebuild.New(opts.Session)
	var resources []resource.Resource

	res, err := svc.ListReports(&codebuild.ListReportsInput{})
	if err != nil {
		return nil, err
	}

	for _, arn := range res.Reports {
		resources = append(resources, &CodeBuildReport{
			svc: svc,
			arn: arn,
		})
	}

	return resources, nil
}

type CodeBuildReport struct {
	svc *codebuild.CodeBuild
	arn *string
}

func (r *CodeBuildReport) Name() string {
	return strings.Split(*r.arn, "report/")[1]
}

func (r *CodeBuildReport) Remove(_ context.Context) error {
	_, err := r.svc.DeleteReport(&codebuild.DeleteReportInput{
		Arn: r.arn,
	})
	return err
}

func (r *CodeBuildReport) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.Name())
	return properties
}

func (r *CodeBuildReport) String() string {
	return r.Name()
}
