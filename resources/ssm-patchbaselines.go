package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ssm" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMPatchBaselineResource = "SSMPatchBaseline"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSMPatchBaselineResource,
		Scope:    nuke.Account,
		Resource: &SSMPatchBaseline{},
		Lister:   &SSMPatchBaselineLister{},
	})
}

type SSMPatchBaselineLister struct{}

func (l *SSMPatchBaselineLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	patchBaselineFilter := []*ssm.PatchOrchestratorFilter{
		{
			Key:    aws.String("OWNER"),
			Values: []*string{aws.String("Self")},
		},
	}

	params := &ssm.DescribePatchBaselinesInput{
		MaxResults: aws.Int64(50),
		Filters:    patchBaselineFilter,
	}

	for {
		output, err := svc.DescribePatchBaselines(params)
		if err != nil {
			return nil, err
		}

		for _, baselineIdentity := range output.BaselineIdentities {
			resources = append(resources, &SSMPatchBaseline{
				svc:             svc,
				ID:              baselineIdentity.BaselineId,
				defaultBaseline: baselineIdentity.DefaultBaseline,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMPatchBaseline struct {
	svc             *ssm.SSM
	ID              *string
	defaultBaseline *bool
}

func (f *SSMPatchBaseline) Remove(_ context.Context) error {
	err := f.DeregisterFromPatchGroups()
	if err != nil {
		return err
	}
	_, err = f.svc.DeletePatchBaseline(&ssm.DeletePatchBaselineInput{
		BaselineId: f.ID,
	})

	return err
}

func (f *SSMPatchBaseline) DeregisterFromPatchGroups() error {
	patchBaseLine, err := f.svc.GetPatchBaseline(&ssm.GetPatchBaselineInput{
		BaselineId: f.ID,
	})
	if err != nil {
		return err
	}
	for _, patchGroup := range patchBaseLine.PatchGroups {
		_, err := f.svc.DeregisterPatchBaselineForPatchGroup(&ssm.DeregisterPatchBaselineForPatchGroupInput{
			BaselineId: f.ID,
			PatchGroup: patchGroup,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *SSMPatchBaseline) String() string {
	return *f.ID
}

func (f *SSMPatchBaseline) Filter() error {
	if *f.defaultBaseline {
		return fmt.Errorf("cannot delete default patch baseline, reassign default first")
	}
	return nil
}
