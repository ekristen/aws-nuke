package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/service/ssm" //nolint:staticcheck
)

type mockSSMAssociationService struct {
	deleteAssociation   func(*ssm.DeleteAssociationInput) (*ssm.DeleteAssociationOutput, error)
	listAssociations    func(*ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error)
	listTagsForResource func(*ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error)
}

func (m *mockSSMAssociationService) DeleteAssociation(
	input *ssm.DeleteAssociationInput,
) (*ssm.DeleteAssociationOutput, error) {
	return m.deleteAssociation(input)
}

func (m *mockSSMAssociationService) ListAssociations(
	input *ssm.ListAssociationsInput,
) (*ssm.ListAssociationsOutput, error) {
	return m.listAssociations(input)
}

func (m *mockSSMAssociationService) ListTagsForResource(
	input *ssm.ListTagsForResourceInput,
) (*ssm.ListTagsForResourceOutput, error) {
	return m.listTagsForResource(input)
}

func Test_Mock_SSMAssociation_ListAndProperties(t *testing.T) {
	tagCalls := 0
	mockSvc := &mockSSMAssociationService{
		listAssociations: func(input *ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error) {
			require.NotNil(t, input.MaxResults)
			assert.Equal(t, int64(50), *input.MaxResults)
			return &ssm.ListAssociationsOutput{
				Associations: []*ssm.Association{
					{
						AssociationId:   ptr.String("association-1"),
						AssociationName: ptr.String("install-agent"),
						InstanceId:      ptr.String("i-0123456789abcdef0"),
						Name:            ptr.String("AWS-RunShellScript"),
						Targets: []*ssm.Target{
							{
								Key: ptr.String("InstanceIds"),
								Values: []*string{
									ptr.String("i-0123456789abcdef0"),
									ptr.String("i-0fedcba9876543210"),
								},
							},
							{
								Key:    ptr.String("tag:Environment"),
								Values: []*string{ptr.String("production")},
							},
						},
					},
					{
						AssociationId: ptr.String("association-2"),
						Name:          ptr.String("AWS-UpdateSSMAgent"),
					},
				},
			}, nil
		},
		listTagsForResource: func(input *ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error) {
			tagCalls++
			assert.Equal(t, ssm.ResourceTypeForTaggingAssociation, ptr.ToString(input.ResourceType))

			if ptr.ToString(input.ResourceId) == "association-1" {
				return &ssm.ListTagsForResourceOutput{
					TagList: []*ssm.Tag{
						{
							Key:   ptr.String("Environment"),
							Value: ptr.String("production"),
						},
					},
				}, nil
			}

			return &ssm.ListTagsForResourceOutput{}, nil
		},
	}

	lister := &SSMAssociationLister{mockSvc: mockSvc}
	resources, err := lister.List(context.TODO(), testListerOpts)
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, 2, tagCalls)

	first := resources[0].(*SSMAssociation)
	properties := first.Properties()
	assert.Equal(t, "association-1", properties.Get("AssociationId"))
	assert.Equal(t, "install-agent", properties.Get("AssociationName"))
	assert.Equal(t, "AWS-RunShellScript", properties.Get("DocumentName"))
	assert.Equal(t, "i-0123456789abcdef0", properties.Get("InstanceId"))
	assert.Equal(
		t,
		"i-0123456789abcdef0,i-0fedcba9876543210",
		properties.Get("Target:InstanceIds"),
	)
	assert.Equal(
		t,
		"i-0123456789abcdef0,i-0fedcba9876543210",
		properties.Get("TargetInstanceIds"),
	)
	assert.Equal(
		t,
		"InstanceIds=i-0123456789abcdef0,i-0fedcba9876543210;tag:Environment=production",
		properties.Get("Targets"),
	)
	assert.Equal(t, "production", properties.Get("tag:Environment"))
	assert.Equal(t, "association-1", first.String())

	second := resources[1].(*SSMAssociation)
	assert.Equal(t, "", second.Properties().Get("AssociationName"))
	assert.Equal(t, "", second.Properties().Get("tag:Environment"))
}

func Test_Mock_SSMAssociation_List_TagErrorSkipsAssociation(t *testing.T) {
	expectedErr := errors.New("unable to list association tags")
	mockSvc := &mockSSMAssociationService{
		listAssociations: func(*ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error) {
			return &ssm.ListAssociationsOutput{
				Associations: []*ssm.Association{
					{AssociationId: ptr.String("association-1")},
					{
						AssociationId:   ptr.String("association-2"),
						AssociationName: ptr.String("retained-association"),
					},
				},
			}, nil
		},
		listTagsForResource: func(input *ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error) {
			if ptr.ToString(input.ResourceId) == "association-1" {
				return nil, expectedErr
			}

			return &ssm.ListTagsForResourceOutput{
				TagList: []*ssm.Tag{
					{Key: ptr.String("Environment"), Value: ptr.String("test")},
				},
			}, nil
		},
	}

	lister := &SSMAssociationLister{mockSvc: mockSvc}
	resources, err := lister.List(context.TODO(), testListerOpts)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	association := resources[0].(*SSMAssociation)
	assert.Equal(t, "association-2", association.String())
	assert.Equal(t, "retained-association", association.Properties().Get("AssociationName"))
	assert.Equal(t, "test", association.Properties().Get("tag:Environment"))
}

func Test_Mock_SSMAssociation_Properties_NilSafe(t *testing.T) {
	association := &SSMAssociation{
		AssociationID: ptr.String("association-1"),
		Targets: []*ssm.Target{
			nil,
			{},
			{Key: ptr.String("InstanceIds"), Values: []*string{nil}},
		},
		TargetMaps: []map[string][]*string{
			nil,
			{"": {nil}},
		},
		Tags: []*ssm.Tag{nil, {}},
	}

	assert.NotPanics(t, func() {
		properties := association.Properties()
		assert.Equal(t, "association-1", properties.Get("AssociationId"))
		assert.Equal(t, "", properties.Get("AssociationName"))
		assert.Equal(t, "", properties.Get("DocumentName"))
		assert.Equal(t, "", properties.Get("tag:"))
	})
}

func Test_Mock_SSMAssociation_Properties_TargetMaps(t *testing.T) {
	association := &SSMAssociation{
		TargetMaps: []map[string][]*string{
			{
				"InstanceIds": {ptr.String("i-0123456789abcdef0")},
				"Operation":   {ptr.String("Install")},
			},
			{
				"InstanceIds": {ptr.String("i-0fedcba9876543210"), nil},
				"":            {ptr.String("ignored")},
			},
		},
	}

	properties := association.Properties()
	assert.Equal(
		t,
		"i-0123456789abcdef0,i-0fedcba9876543210",
		properties.Get("TargetMap:InstanceIds"),
	)
	assert.Equal(t, "Install", properties.Get("TargetMap:Operation"))
	assert.Equal(
		t,
		"InstanceIds=i-0123456789abcdef0,i-0fedcba9876543210;Operation=Install",
		properties.Get("TargetMaps"),
	)
	assert.Equal(
		t,
		"i-0123456789abcdef0,i-0fedcba9876543210",
		properties.Get("TargetInstanceIds"),
	)
	assert.Empty(t, properties.Get("Targets"))
}

func Test_Mock_SSMAssociation_List_Pagination(t *testing.T) {
	listCalls := 0
	taggedAssociationIDs := make([]string, 0, 2)
	mockSvc := &mockSSMAssociationService{
		listAssociations: func(input *ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error) {
			listCalls++
			require.NotNil(t, input.MaxResults)
			assert.Equal(t, int64(50), *input.MaxResults)

			switch listCalls {
			case 1:
				assert.Nil(t, input.NextToken)
				return &ssm.ListAssociationsOutput{
					Associations: []*ssm.Association{
						{AssociationId: ptr.String("association-1")},
					},
					NextToken: ptr.String("next-page"),
				}, nil
			case 2:
				assert.Equal(t, "next-page", ptr.ToString(input.NextToken))
				return &ssm.ListAssociationsOutput{
					Associations: []*ssm.Association{
						{AssociationId: ptr.String("association-2")},
					},
				}, nil
			default:
				t.Fatalf("unexpected ListAssociations call %d", listCalls)
				return &ssm.ListAssociationsOutput{}, nil
			}
		},
		listTagsForResource: func(input *ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error) {
			taggedAssociationIDs = append(taggedAssociationIDs, ptr.ToString(input.ResourceId))
			return &ssm.ListTagsForResourceOutput{}, nil
		},
	}

	lister := &SSMAssociationLister{mockSvc: mockSvc}
	resources, err := lister.List(context.TODO(), testListerOpts)
	require.NoError(t, err)
	require.Len(t, resources, 2)
	assert.Equal(t, 2, listCalls)
	assert.Equal(t, []string{"association-1", "association-2"}, taggedAssociationIDs)
	assert.Equal(t, "association-1", resources[0].(*SSMAssociation).String())
	assert.Equal(t, "association-2", resources[1].(*SSMAssociation).String())
}

func Test_Mock_SSMAssociation_List_NilAssociationID(t *testing.T) {
	mockSvc := &mockSSMAssociationService{
		listAssociations: func(*ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error) {
			return &ssm.ListAssociationsOutput{
				Associations: []*ssm.Association{{}},
			}, nil
		},
		listTagsForResource: func(*ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error) {
			t.Fatal("ListTagsForResource must not be called without an association ID")
			return &ssm.ListTagsForResourceOutput{}, nil
		},
	}

	lister := &SSMAssociationLister{mockSvc: mockSvc}
	resources, err := lister.List(context.TODO(), testListerOpts)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Empty(t, resources[0].(*SSMAssociation).String())
	assert.Empty(t, resources[0].(*SSMAssociation).Properties().Get("AssociationId"))
}

func Test_Mock_SSMAssociation_List_NilTagOutput(t *testing.T) {
	mockSvc := &mockSSMAssociationService{
		listAssociations: func(*ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error) {
			return &ssm.ListAssociationsOutput{
				Associations: []*ssm.Association{
					{AssociationId: ptr.String("association-1")},
				},
			}, nil
		},
		listTagsForResource: func(*ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error) {
			return nil, nil //nolint:nilnil // Exercise defensive handling of a successful response with no output.
		},
	}

	lister := &SSMAssociationLister{mockSvc: mockSvc}
	resources, err := lister.List(context.TODO(), testListerOpts)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Empty(t, resources[0].(*SSMAssociation).Properties().Get("tag:Environment"))
}

func Test_Mock_SSMAssociation_Remove(t *testing.T) {
	mockSvc := &mockSSMAssociationService{
		deleteAssociation: func(input *ssm.DeleteAssociationInput) (*ssm.DeleteAssociationOutput, error) {
			assert.Equal(t, "association-1", ptr.ToString(input.AssociationId))
			return &ssm.DeleteAssociationOutput{}, nil
		},
	}
	association := &SSMAssociation{
		svc:           mockSvc,
		AssociationID: ptr.String("association-1"),
	}

	assert.NoError(t, association.Remove(context.TODO()))
}
