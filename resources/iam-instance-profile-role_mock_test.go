package resources

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_iamiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_IAMInstanceProfileRole_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamInstanceProfileRole := IAMInstanceProfileRole{
		svc: mockIAM,
		role: &iam.Role{
			RoleName: ptr.String("role:foobar"),
		},
		profile: &iam.InstanceProfile{
			Arn:                 ptr.String("arn:aws:iam::123456789012:instance-profile/profile:foobar"),
			InstanceProfileName: ptr.String("profile:foobar"),
			CreateDate:          ptr.Time(time.Now()),
			Roles: []*iam.Role{
				{
					Arn:      ptr.String("arn:aws:iam::123456789012:role/role:foobar"),
					RoleName: ptr.String("role:foobar"),
				},
			},
		},
	}

	mockIAM.EXPECT().ListInstanceProfiles(gomock.Any()).Return(&iam.ListInstanceProfilesOutput{
		InstanceProfiles: []*iam.InstanceProfile{
			iamInstanceProfileRole.profile,
		},
		IsTruncated: ptr.Bool(false),
	}, nil)

	mockIAM.EXPECT().GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: iamInstanceProfileRole.profile.InstanceProfileName,
	}).Return(&iam.GetInstanceProfileOutput{
		InstanceProfile: iamInstanceProfileRole.profile,
	}, nil)

	lister := IAMInstanceProfileRoleLister{
		mockSvc: mockIAM,
	}

	resources, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region: &nuke.Region{
			Name: "us-east-2",
		},
		Session: session.Must(session.NewSession()),
	})
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_IAMInstanceProfileRole_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)

	iamInstanceProfileRole := IAMInstanceProfileRole{
		svc: mockIAM,
		role: &iam.Role{
			RoleName: aws.String("role:foobar"),
		},
		profile: &iam.InstanceProfile{
			InstanceProfileName: aws.String("profile:foobar"),
		},
	}

	mockIAM.EXPECT().RemoveRoleFromInstanceProfile(gomock.Eq(&iam.RemoveRoleFromInstanceProfileInput{
		RoleName:            iamInstanceProfileRole.role.RoleName,
		InstanceProfileName: iamInstanceProfileRole.profile.InstanceProfileName,
	})).Return(&iam.RemoveRoleFromInstanceProfileOutput{}, nil)

	err := iamInstanceProfileRole.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_IAMInstanceProfileRole_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now()

	iamInstanceProfileRole := IAMInstanceProfileRole{
		role: &iam.Role{
			Arn:        ptr.String("arn:aws:iam::123456789012:role/role:foobar"),
			RoleName:   ptr.String("role:foobar"),
			Path:       ptr.String("/"),
			CreateDate: ptr.Time(now),
			Tags: []*iam.Tag{
				{
					Key:   ptr.String("Name"),
					Value: ptr.String("bar"),
				},
			},
		},
		profile: &iam.InstanceProfile{
			InstanceProfileName: ptr.String("profile:foobar"),
			Tags: []*iam.Tag{
				{
					Key:   ptr.String("Name"),
					Value: ptr.String("foo"),
				},
			},
		},
	}

	props := iamInstanceProfileRole.Properties()
	a.Equal("profile:foobar", props.Get("InstanceProfile"))
	a.Equal("role:foobar", props.Get("InstanceRole"))
	a.Equal("/", props.Get("role:Path"))
	a.Equal(now.Format(time.RFC3339), props.Get("role:CreateDate"))
	a.Equal("foo", props.Get("tag:Name"))
	a.Equal("bar", props.Get("tag:role:Name"))

	a.Equal("profile:foobar -> role:foobar", iamInstanceProfileRole.String())
}
