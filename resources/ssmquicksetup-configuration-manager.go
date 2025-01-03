package resources

import (
	"context"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/ssmquicksetup"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMQuickSetupConfigurationManagerResource = "SSMQuickSetupConfigurationManager"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSMQuickSetupConfigurationManagerResource,
		Scope:    nuke.Account,
		Resource: &SSMQuickSetupConfigurationManager{},
		Lister:   &SSMQuickSetupConfigurationManagerLister{},
	})
}

type SSMQuickSetupConfigurationManagerLister struct{}

func (l *SSMQuickSetupConfigurationManagerLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ssmquicksetup.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListConfigurationManagers(ctx, &ssmquicksetup.ListConfigurationManagersInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.ConfigurationManagersList {
		resources = append(resources, &SSMQuickSetupConfigurationManager{
			svc:  svc,
			ARN:  p.ManagerArn,
			Name: p.Name,
		})
	}

	return resources, nil
}

type SSMQuickSetupConfigurationManager struct {
	svc  *ssmquicksetup.Client
	ARN  *string
	Name *string
}

// GetName returns the name of the resource or the last part of the ARN if not set so that the stringer resource has
// a value to display
func (r *SSMQuickSetupConfigurationManager) GetName() string {
	if ptr.ToString(r.Name) != "" {
		return ptr.ToString(r.Name)
	}

	parts := strings.Split(ptr.ToString(r.ARN), "/")
	return parts[len(parts)-1]
}

func (r *SSMQuickSetupConfigurationManager) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteConfigurationManager(ctx, &ssmquicksetup.DeleteConfigurationManagerInput{
		ManagerArn: r.ARN,
	})
	return err
}

func (r *SSMQuickSetupConfigurationManager) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SSMQuickSetupConfigurationManager) String() string {
	return r.GetName()
}
