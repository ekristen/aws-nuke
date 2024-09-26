package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/quicksight/quicksightiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const QuickSightUserResource = "QuickSightUser"

func init() {
	registry.Register(&registry.Registration{
		Name:   QuickSightUserResource,
		Scope:  nuke.Account,
		Lister: &QuickSightUserLister{},
	})
}

type QuickSightUserLister struct {
	stsService        stsiface.STSAPI
	quicksightService quicksightiface.QuickSightAPI
}

func (l *QuickSightUserLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	var stsSvc stsiface.STSAPI
	if l.stsService != nil {
		stsSvc = l.stsService
	} else {
		stsSvc = sts.New(opts.Session)
	}

	var quicksightSvc quicksightiface.QuickSightAPI
	if l.quicksightService != nil {
		quicksightSvc = l.quicksightService
	} else {
		quicksightSvc = quicksight.New(opts.Session)
	}

	callerID, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	accountID := callerID.Account

	// TODO: support all namespaces
	namespace := ptr.String("default")

	err = quicksightSvc.ListUsersPages(&quicksight.ListUsersInput{
		AwsAccountId: accountID,
		Namespace:    namespace,
	}, func(output *quicksight.ListUsersOutput, lastPage bool) bool {
		for _, user := range output.UserList {
			resources = append(resources, &QuickSightUser{
				svc:         quicksightSvc,
				accountID:   accountID,
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
		return nil, err
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
