package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OpsWorksUserProfileResource = "OpsWorksUserProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:   OpsWorksUserProfileResource,
		Scope:  nuke.Account,
		Lister: &OpsWorksUserProfileLister{},
	})
}

type OpsWorksUserProfileLister struct{}

func (l *OpsWorksUserProfileLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworks.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// TODO: pass in account information via ListerOpts to avoid additional calls.

	identityOutput, err := sts.New(opts.Session).GetCallerIdentity(nil)
	if err != nil {
		return nil, err
	}

	params := &opsworks.DescribeUserProfilesInput{}

	output, err := svc.DescribeUserProfiles(params)
	if err != nil {
		return nil, err
	}

	for _, userProfile := range output.UserProfiles {
		resources = append(resources, &OpsWorksUserProfile{
			svc:        svc,
			callingArn: identityOutput.Arn,
			ARN:        userProfile.IamUserArn,
		})
	}

	return resources, nil
}

type OpsWorksUserProfile struct {
	svc        *opsworks.OpsWorks
	ARN        *string
	callingArn *string
}

func (f *OpsWorksUserProfile) Filter() error {
	if *f.callingArn == *f.ARN {
		return fmt.Errorf("cannot delete OpsWorksUserProfile of calling User")
	}
	return nil
}

func (f *OpsWorksUserProfile) Remove(_ context.Context) error {
	_, err := f.svc.DeleteUserProfile(&opsworks.DeleteUserProfileInput{
		IamUserArn: f.ARN,
	})

	return err
}

func (f *OpsWorksUserProfile) String() string {
	return *f.ARN
}
