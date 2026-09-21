package resources

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"                //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/quicksight" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_quicksightiface"
)

func Test_Mock_QuickSightIdentityRegion(t *testing.T) {
	accountID := testListerOpts.AccountID

	cases := []struct {
		name     string
		output   *quicksight.ListNamespacesOutput
		err      error
		expected string
	}{
		{
			name: "default namespace",
			output: &quicksight.ListNamespacesOutput{
				Namespaces: []*quicksight.NamespaceInfoV2{
					{Name: aws.String("other"), CapacityRegion: aws.String("eu-west-1")},
					{Name: aws.String("default"), CapacityRegion: aws.String("us-east-1")},
				},
			},
			expected: "us-east-1",
		},
		{
			name: "without a default namespace",
			output: &quicksight.ListNamespacesOutput{
				Namespaces: []*quicksight.NamespaceInfoV2{
					{Name: aws.String("other"), CapacityRegion: aws.String("eu-west-1")},
				},
			},
			expected: "eu-west-1",
		},
		{
			name:     "no namespaces",
			output:   &quicksight.ListNamespacesOutput{},
			expected: "",
		},
		{
			name: "rejected outside the identity region",
			err: &quicksight.AccessDeniedException{
				Message_: aws.String("Operation is being called from endpoint us-east-2, " +
					"but your identity region is us-east-1. Please use the us-east-1 endpoint."),
			},
			expected: "us-east-1",
		},
		{
			name:     "unrelated error",
			err:      errors.New("MOCK_ERROR"),
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertions := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQuickSightAPI := mock_quicksightiface.NewMockQuickSightAPI(ctrl)
			mockQuickSightAPI.EXPECT().ListNamespaces(&quicksight.ListNamespacesInput{
				AwsAccountId: accountID,
			}).Return(tc.output, tc.err)

			assertions.Equal(tc.expected, QuickSightIdentityRegion(mockQuickSightAPI, accountID))
		})
	}
}
