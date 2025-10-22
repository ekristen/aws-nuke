package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	inspectortypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Inspector2Resource = "Inspector2"

func init() {
	registry.Register(&registry.Registration{
		Name:     Inspector2Resource,
		Scope:    nuke.Account,
		Resource: &Inspector2{},
		Lister:   &Inspector2Lister{},
	})
}

type Inspector2Lister struct{}

func (l *Inspector2Lister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := inspector2.NewFromConfig(*opts.Config)

	resources := make([]resource.Resource, 0)

	resp, err := svc.BatchGetAccountStatus(ctx, &inspector2.BatchGetAccountStatusInput{})
	if err != nil {
		return resources, err
	}

	for _, a := range resp.Accounts {
		if a.State.Status != inspectortypes.StatusDisabled {
			resources = append(resources, &Inspector2{
				svc:       svc,
				AccountID: a.AccountId,
				Status:    &a.State.Status,
				ResourceState: map[string]string{
					string(inspectortypes.ResourceScanTypeEc2):            string(a.ResourceState.Ec2.Status),
					string(inspectortypes.ResourceScanTypeEcr):            string(a.ResourceState.Ecr.Status),
					string(inspectortypes.ResourceScanTypeLambda):         string(a.ResourceState.Lambda.Status),
					string(inspectortypes.ResourceScanTypeLambdaCode):     string(a.ResourceState.LambdaCode.Status),
					string(inspectortypes.ResourceScanTypeCodeRepository): string(a.ResourceState.CodeRepository.Status),
				},
			})
		}
	}

	return resources, nil
}

type Inspector2 struct {
	svc           *inspector2.Client
	triggered     bool
	AccountID     *string
	Status        *inspectortypes.Status
	ResourceState map[string]string `property:"tagPrefix=resourceType"`
}

func (e *Inspector2) GetEnabledResources() []inspectortypes.ResourceScanType {
	var resources = make([]inspectortypes.ResourceScanType, 0)
	for k, v := range e.ResourceState {
		if v == string(inspectortypes.StatusEnabled) {
			resources = append(resources, inspectortypes.ResourceScanType(k))
		}
	}
	return resources
}

func (e *Inspector2) Remove(ctx context.Context) error {
	enabledResources := e.GetEnabledResources()
	if len(enabledResources) == 0 {
		return nil
	}

	_, err := e.svc.Disable(ctx, &inspector2.DisableInput{
		AccountIds:    []string{*e.AccountID},
		ResourceTypes: enabledResources,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *Inspector2) String() string {
	return *e.AccountID
}

func (e *Inspector2) Properties() types.Properties {
	return types.NewPropertiesFromStruct(e)
}
