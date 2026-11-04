package resources

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/ekristen/libnuke/pkg/registry"
)

type mockECSDaemonClient struct {
	mock.Mock
}

func (m *mockECSDaemonClient) ListDaemons(ctx context.Context, params *ecs.ListDaemonsInput,
	_ ...func(*ecs.Options)) (*ecs.ListDaemonsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.ListDaemonsOutput), args.Error(1)
}

func (m *mockECSDaemonClient) ListTagsForResource(ctx context.Context, params *ecs.ListTagsForResourceInput,
	_ ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.ListTagsForResourceOutput), args.Error(1)
}

func (m *mockECSDaemonClient) DeleteDaemon(ctx context.Context, params *ecs.DeleteDaemonInput,
	_ ...func(*ecs.Options)) (*ecs.DeleteDaemonOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.DeleteDaemonOutput), args.Error(1)
}

func Test_Mock_ECSDaemon_List(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockECSDaemonClient)

	daemon1 := "arn:aws:ecs:us-east-2:123456789012:daemon/cluster-name/daemon-1"
	daemon2 := "arn:aws:ecs:us-east-2:123456789012:daemon/cluster-name/daemon-2"
	created := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	mockSvc.On("ListDaemons", mock.Anything, &ecs.ListDaemonsInput{
		MaxResults: ptr.Int32(100),
	}).Return(&ecs.ListDaemonsOutput{
		DaemonSummariesList: []ecstypes.DaemonSummary{
			{DaemonArn: ptr.String(daemon1), Status: ecstypes.DaemonStatusActive, CreatedAt: &created},
		},
		NextToken: ptr.String("next"),
	}, nil).Once()

	mockSvc.On("ListDaemons", mock.Anything, &ecs.ListDaemonsInput{
		MaxResults: ptr.Int32(100),
		NextToken:  ptr.String("next"),
	}).Return(&ecs.ListDaemonsOutput{
		DaemonSummariesList: []ecstypes.DaemonSummary{
			{DaemonArn: ptr.String(daemon2), Status: ecstypes.DaemonStatusDeleteInProgress},
		},
	}, nil).Once()

	mockSvc.On("ListTagsForResource", mock.Anything, &ecs.ListTagsForResourceInput{
		ResourceArn: ptr.String(daemon1),
	}).Return(&ecs.ListTagsForResourceOutput{
		Tags: []ecstypes.Tag{{Key: ptr.String("Environment"), Value: ptr.String("nonprod")}},
	}, nil)

	mockSvc.On("ListTagsForResource", mock.Anything, &ecs.ListTagsForResourceInput{
		ResourceArn: ptr.String(daemon2),
	}).Return((*ecs.ListTagsForResourceOutput)(nil), errors.New("access denied"))

	lister := &ECSDaemonLister{mockSvc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	r1 := resources[0].(*ECSDaemon)
	a.Equal(daemon1, r1.String())
	a.Equal("ACTIVE", *r1.Status)
	a.Equal(created, *r1.CreatedAt)
	a.Equal("nonprod", r1.Tags["Environment"])
	a.Nil(r1.Filter())

	r2 := resources[1].(*ECSDaemon)
	a.Equal(daemon2, r2.String())
	a.Empty(r2.Tags)
	a.EqualError(r2.Filter(), "already being deleted")

	mockSvc.AssertExpectations(t)
}

func Test_Mock_ECSDaemon_ListError(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockECSDaemonClient)
	mockSvc.On("ListDaemons", mock.Anything, mock.Anything).
		Return((*ecs.ListDaemonsOutput)(nil), errors.New("boom"))

	lister := &ECSDaemonLister{mockSvc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.EqualError(err, "boom")
	a.Nil(resources)
}

func Test_Mock_ECSDaemon_Remove(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockECSDaemonClient)
	daemonArn := "arn:aws:ecs:us-east-2:123456789012:daemon/cluster-name/daemon-1"

	mockSvc.On("DeleteDaemon", mock.Anything, &ecs.DeleteDaemonInput{
		DaemonArn: ptr.String(daemonArn),
	}).Return(&ecs.DeleteDaemonOutput{}, nil)

	r := &ECSDaemon{
		svc: mockSvc,
		ARN: ptr.String(daemonArn),
	}

	a.Nil(r.Remove(context.TODO()))
	mockSvc.AssertExpectations(t)
}

func Test_Mock_ECSDaemon_Properties(t *testing.T) {
	a := assert.New(t)

	r := &ECSDaemon{
		ARN:    ptr.String("arn:aws:ecs:us-east-2:123456789012:daemon/cluster-name/daemon-1"),
		Status: ptr.String("ACTIVE"),
		Tags:   map[string]string{"Owner": "agent"},
	}

	props := r.Properties()
	a.Equal("arn:aws:ecs:us-east-2:123456789012:daemon/cluster-name/daemon-1", props.Get("ARN"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("agent", props.Get("tag:Owner"))
}

// The capacity provider cannot be deleted while a daemon is still associated with it,
// so daemons must be removed first.
func Test_ECSCapacityProvider_DependsOnECSDaemon(t *testing.T) {
	a := assert.New(t)

	reg := registry.GetRegistration(ECSCapacityProviderResource)
	a.NotNil(reg)
	a.Contains(reg.DependsOn, ECSDaemonResource)
}
