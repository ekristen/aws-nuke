package resources

import (
	"context"
	"errors"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/quicksight/quicksightiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const QuickSightUserResource = "QuickSightUser"

func init() {
	registry.Register(&registry.Registration{
		Name:     QuickSightUserResource,
		Scope:    nuke.Account,
		Resource: &QuickSightUserLister{},
		Lister:   &QuickSightUserLister{},
	})
}

type QuickSightUserLister struct {
	quicksightService quicksightiface.QuickSightAPI
}

func (l *QuickSightUserLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	var quicksightSvc quicksightiface.QuickSightAPI
	if l.quicksightService != nil {
		quicksightSvc = l.quicksightService
	} else {
		quicksightSvc = quicksight.New(opts.Session)
	}

	// TODO: support all namespaces
	namespace := ptr.String("default")

	err := quicksightSvc.ListUsersPages(&quicksight.ListUsersInput{
		AwsAccountId: opts.AccountID,
		Namespace:    namespace,
	}, func(output *quicksight.ListUsersOutput, lastPage bool) bool {
		for _, user := range output.UserList {
			resources = append(resources, &QuickSightUser{
				svc:         quicksightSvc,
				accountID:   opts.AccountID,
				PrincipalID: user.PrincipalId,
				UserName:    user.UserName,
				Active:      user.Active,
				Role:        user.Role,
				Namespace:   namespace,
			})
		}
		return !lastPage
	})
	if err != nil {
		var notFoundException *quicksight.ResourceNotFoundException
		if !errors.As(err, &notFoundException) {
			return nil, err
		}
		return resources, nil
	}

	return resources, nil
}

type QuickSightUser struct {
	svc         quicksightiface.QuickSightAPI
	accountID   *string
	UserName    *string
	PrincipalID *string
	Active      *bool
	Role        *string
	Namespace   *string
}

func (r *QuickSightUser) Remove(_ context.Context) error {
	_, err := r.svc.DeleteUserByPrincipalId(&quicksight.DeleteUserByPrincipalIdInput{
		AwsAccountId: r.accountID,
		Namespace:    r.Namespace,
		PrincipalId:  r.PrincipalID,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *QuickSightUser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *QuickSightUser) String() string {
	return *r.PrincipalID
}
