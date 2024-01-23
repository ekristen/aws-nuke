package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const TransferServerUserResource = "TransferServerUser"

func init() {
	resource.Register(&resource.Registration{
		Name:   TransferServerUserResource,
		Scope:  nuke.Account,
		Lister: &TransferServerUserLister{},
	})
}

type TransferServerUserLister struct{}

func (l *TransferServerUserLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := transfer.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &transfer.ListServersInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListServers(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Servers {
			userParams := &transfer.ListUsersInput{
				MaxResults: aws.Int64(100),
				ServerId:   item.ServerId,
			}

			for {
				userOutput, err := svc.ListUsers(userParams)
				if err != nil {
					return nil, err
				}

				for _, user := range userOutput.Users {
					descOutput, err := svc.DescribeUser(&transfer.DescribeUserInput{
						ServerId: item.ServerId,
						UserName: user.UserName,
					})
					if err != nil {
						return nil, err
					}

					resources = append(resources, &TransferServerUser{
						svc:      svc,
						username: user.UserName,
						serverID: item.ServerId,
						tags:     descOutput.User.Tags,
					})

				}

				if userOutput.NextToken == nil {
					break
				}

				userParams.NextToken = userOutput.NextToken
			}
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type TransferServerUser struct {
	svc      *transfer.Transfer
	username *string
	serverID *string
	tags     []*transfer.Tag
}

func (ts *TransferServerUser) Remove(_ context.Context) error {
	_, err := ts.svc.DeleteUser(&transfer.DeleteUserInput{
		ServerId: ts.serverID,
		UserName: ts.username,
	})

	return err
}

func (ts *TransferServerUser) String() string {
	return fmt.Sprintf("%s -> %s", *ts.serverID, *ts.username)
}

func (ts *TransferServerUser) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range ts.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.
		Set("Username", ts.username).
		Set("ServerID", ts.serverID)
	return properties
}
