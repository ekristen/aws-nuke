package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector2"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const Inspector2Resource = "Inspector2"

func init() {
	resource.Register(resource.Registration{
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
				accountId: a.AccountId,
			})
		}
	}

	return resources, nil
}

type Inspector2 struct {
	svc       *inspector2.Inspector2
	accountId *string
}

func (e *Inspector2) Remove(_ context.Context) error {
	_, err := e.svc.Disable(&inspector2.DisableInput{
		AccountIds:    []*string{e.accountId},
		ResourceTypes: aws.StringSlice(inspector2.ResourceScanType_Values()),
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *Inspector2) String() string {
	return *e.accountId
}
