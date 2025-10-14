package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/macie2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MacieResource = "Macie"

func init() {
	registry.Register(&registry.Registration{
		Name:     MacieResource,
		Scope:    nuke.Account,
		Resource: &Macie{},
		Lister:   &MacieLister{},
	})
}

type MacieLister struct{}

func (l *MacieLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := macie2.New(opts.Session)

	status, err := svc.GetMacieSession(&macie2.GetMacieSessionInput{})
	if err != nil {
		if err.Error() == "AccessDeniedException: Macie is not enabled" {
			return nil, nil
		} else {
			return nil, err
		}
	}

	resources := make([]resource.Resource, 0)
	if *status.Status == macie2.AdminStatusEnabled {
		resources = append(resources, &Macie{
			svc: svc,
		})
	}

	return resources, nil
}

type Macie struct {
	svc *macie2.Macie2
}

func (b *Macie) Remove(_ context.Context) error {
	_, err := b.svc.DisableMacie(&macie2.DisableMacieInput{})
	if err != nil {
		return err
	}
	return nil
}

func (b *Macie) String() string {
	return "macie"
}
