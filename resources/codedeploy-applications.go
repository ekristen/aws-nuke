package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/codedeploy"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeDeployApplicationResource = "CodeDeployApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:   CodeDeployApplicationResource,
		Scope:  nuke.Account,
		Lister: &CodeDeployApplicationLister{},
	})
}

type CodeDeployApplicationLister struct{}

func (l *CodeDeployApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codedeploy.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codedeploy.ListApplicationsInput{}

	for {
		resp, err := svc.ListApplications(params)
		if err != nil {
			return nil, err
		}

		for _, application := range resp.Applications {
			resources = append(resources, &CodeDeployApplication{
				svc:             svc,
				applicationName: application,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CodeDeployApplication struct {
	svc             *codedeploy.CodeDeploy
	applicationName *string
}

func (f *CodeDeployApplication) Remove(_ context.Context) error {
	_, err := f.svc.DeleteApplication(&codedeploy.DeleteApplicationInput{
		ApplicationName: f.applicationName,
	})

	return err
}

func (f *CodeDeployApplication) String() string {
	return *f.applicationName
}
