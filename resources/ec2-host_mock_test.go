package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_ec2host"
)

func Test_Mock_EC2Host_Remove(t *testing.T) {
	tests := []struct {
		name        string
		output      *ec2.ReleaseHostsOutput
		apiErr      error
		expectedErr string
	}{
		{
			name:   "successful release",
			output: &ec2.ReleaseHostsOutput{Successful: []string{"h-1234567890"}},
		},
		{
			name:        "API error",
			output:      nil,
			apiErr:      errors.New("request failed"),
			expectedErr: "request failed",
		},
		{
			name: "unsuccessful release",
			output: &ec2.ReleaseHostsOutput{
				Unsuccessful: []ec2types.UnsuccessfulItem{
					{
						ResourceId: ptr.String("h-1234567890"),
						Error: &ec2types.UnsuccessfulItemError{
							Code:    ptr.String("Client.InvalidHost.Occupied"),
							Message: ptr.String("The host is occupied."),
						},
					},
				},
			},
			expectedErr: "failed to release EC2 host \"h-1234567890\": Client.InvalidHost.Occupied: The host is occupied.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockSvc := mock_ec2host.NewMockEC2HostAPI(ctrl)
			host := &EC2Host{
				svc: mockSvc,
				host: &ec2types.Host{
					HostId: ptr.String("h-1234567890"),
				},
			}

			mockSvc.EXPECT().
				ReleaseHosts(gomock.Any(), &ec2.ReleaseHostsInput{
					HostIds: []string{"h-1234567890"},
				}).
				Return(test.output, test.apiErr)

			err := host.Remove(context.Background())
			if test.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedErr)
			}
		})
	}
}
