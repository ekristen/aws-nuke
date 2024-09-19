package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/quicksight/quicksightiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
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

type QuickSightUser struct {
	svc         quicksightiface.QuickSightAPI
	accountID   *string
	principalID *string
}

func (l *QuickSightUserLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

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

	var resources []resource.Resource

	err = quicksightSvc.ListUsersPages(&quicksight.ListUsersInput{
		AwsAccountId: accountID,
		Namespace:    aws.String("default"),
	}, func(output *quicksight.ListUsersOutput, lastPage bool) bool {
		for _, user := range output.UserList {
			resources = append(resources, &QuickSightUser{
				svc:         quicksightSvc,
				accountID:   accountID,
				principalID: user.PrincipalId,
			})
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *QuickSightUser) Remove(_ context.Context) error {
	_, err := r.svc.DeleteUserByPrincipalId(&quicksight.DeleteUserByPrincipalIdInput{
		AwsAccountId: r.accountID,
		Namespace:    aws.String("default"),
		PrincipalId:  r.principalID,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *QuickSightUser) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("PrincipalId", r.principalID)

	return properties
}

func (r *QuickSightUser) String() string {
	return *r.principalID
}

func (r *QuickSightUser) Filter() error {
	return nil
}
