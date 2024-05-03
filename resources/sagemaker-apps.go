package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerAppResource = "SageMakerApp"

func init() {
	registry.Register(&registry.Registration{
		Name:   SageMakerAppResource,
		Scope:  nuke.Account,
		Lister: &SageMakerAppLister{},
	})
}

type SageMakerAppLister struct{}

func (l *SageMakerAppLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListAppsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListApps(params)
		if err != nil {
			return nil, err
		}

		for _, app := range resp.Apps {
			resources = append(resources, &SageMakerApp{
				svc:             svc,
				domainID:        app.DomainId,
				appName:         app.AppName,
				appType:         app.AppType,
				userProfileName: app.UserProfileName,
				spaceName:       app.SpaceName,
				status:          app.Status,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerApp struct {
	svc             *sagemaker.SageMaker
	domainID        *string
	appName         *string
	appType         *string
	userProfileName *string
	spaceName       *string
	status          *string
}

func (f *SageMakerApp) Remove(_ context.Context) error {
	_, err := f.svc.DeleteApp(&sagemaker.DeleteAppInput{
		DomainId:        f.domainID,
		AppName:         f.appName,
		AppType:         f.appType,
		UserProfileName: f.userProfileName,
		SpaceName:       f.spaceName,
	})

	return err
}

func (f *SageMakerApp) String() string {
	return *f.appName
}

func (f *SageMakerApp) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("DomainID", f.domainID).
		Set("AppName", f.appName).
		Set("AppType", f.appType).
		Set("UserProfileName", f.userProfileName).
		Set("SpaceName", f.spaceName)
	return properties
}

func (f *SageMakerApp) Filter() error {
	if *f.status == sagemaker.AppStatusDeleted {
		return fmt.Errorf("already deleted")
	}
	return nil
}
