package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const Inspector2Resource = "Inspector2"

func init() {
	registry.Register(&registry.Registration{
		Name:   Inspector2Resource,
		Scope:  nuke.Account,
		Lister: &Inspector2Lister{},
	})
}

type Inspector2Lister struct{}

func (l *Inspector2Lister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := inspector2.New(opts.Session)

	resources := make([]resource.Resource, 0)

	resp, err := svc.BatchGetAccountStatus(nil)
	if err != nil {
		return resources, err
	}
	for _, a := range resp.Accounts {
		if *a.State.Status != inspector2.StatusDisabled {
			resources = append(resources, &Inspector2{
				svc:       svc,
				AccountID: a.AccountId,
				ResourceState: map[string]string{
					inspector2.ResourceScanTypeEc2:        *a.ResourceState.Ec2.Status,
					inspector2.ResourceScanTypeEcr:        *a.ResourceState.Ecr.Status,
					inspector2.ResourceScanTypeLambda:     *a.ResourceState.Lambda.Status,
					inspector2.ResourceScanTypeLambdaCode: *a.ResourceState.LambdaCode.Status,
				},
			})
		}
	}

	return resources, nil
}

type Inspector2 struct {
	svc           *inspector2.Inspector2
	AccountID     *string
	ResourceState map[string]string `property:"tagPrefix=resourceType"`
}

func (e *Inspector2) GetEnabledResources() []string {
	var resources = make([]string, 0)
	for k, v := range e.ResourceState {
		if v == inspector2.StatusEnabled {
			resources = append(resources, k)
		}
	}
	return resources
}

func (e *Inspector2) Remove(_ context.Context) error {
	_, err := e.svc.Disable(&inspector2.DisableInput{
		AccountIds:    []*string{e.AccountID},
		ResourceTypes: aws.StringSlice(e.GetEnabledResources()),
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
