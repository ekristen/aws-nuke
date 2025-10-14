package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
)

func Test_Mock_IAMInstanceProfile_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	instanceProfile := &iam.InstanceProfile{
		Arn:                 ptr.String("arn:aws:iam::123456789012:instance-profile/profile:foobar"),
		InstanceProfileName: ptr.String("profile:foobar"),
		CreateDate:          ptr.Time(time.Now()),
		Roles: []*iam.Role{
			{
				Arn:      ptr.String("arn:aws:iam::123456789012:role/role:foobar"),
				RoleName: ptr.String("role:foobar"),
			},
		},
	}

	instanceProfile2 := &iam.InstanceProfile{
		Arn:                 ptr.String("arn:aws:iam::123456789012:instance-profile/profile:foobar2"),
		InstanceProfileName: ptr.String("profile:foobar2"),
		CreateDate:          ptr.Time(time.Now()),
		Roles: []*iam.Role{
			{
				Arn:      ptr.String("arn:aws:iam::123456789012:role/role:foobar2"),
				RoleName: ptr.String("role:foobar2"),
			},
		},
	}

	mockIAM.EXPECT().ListInstanceProfiles(gomock.Any()).Return(&iam.ListInstanceProfilesOutput{
		InstanceProfiles: []*iam.InstanceProfile{
			instanceProfile,
			instanceProfile2,
		},
	}, nil)

	mockIAM.EXPECT().GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: ptr.String("profile:foobar"),
	}).Return(&iam.GetInstanceProfileOutput{
		InstanceProfile: instanceProfile,
	}, nil)

	mockIAM.EXPECT().GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: ptr.String("profile:foobar2"),
	}).Return(nil, awserr.New("400", "InstanceProfileNotFound", nil))

	lister := IAMInstanceProfileLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), testListerOpts)

	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_IAMInstanceProfile_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamInstanceProfile := IAMInstanceProfile{
		svc:  mockIAM,
		Name: ptr.String("ip:foobar"),
		Path: ptr.String("/"),
	}

	mockIAM.EXPECT().DeleteInstanceProfile(gomock.Eq(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: iamInstanceProfile.Name,
	})).Return(&iam.DeleteInstanceProfileOutput{}, nil)

	err := iamInstanceProfile.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMInstanceProfile_Properties(t *testing.T) {
	a := assert.New(t)

	iamInstanceProfile := IAMInstanceProfile{
		Name: ptr.String("ip:foobar"),
		Path: ptr.String("/"),
		Tags: []*iam.Tag{
			{
				Key:   ptr.String("foo"),
				Value: ptr.String("bar"),
			},
		},
	}

	a.Equal("ip:foobar", iamInstanceProfile.Properties().Get("Name"))
	a.Equal("/", iamInstanceProfile.Properties().Get("Path"))
	a.Equal("bar", iamInstanceProfile.Properties().Get("tag:foo"))

	a.Equal("ip:foobar", iamInstanceProfile.String())
}
