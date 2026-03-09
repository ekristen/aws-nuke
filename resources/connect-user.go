package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectUserResource = "ConnectUser"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectUserResource,
		Scope:    nuke.Account,
		Resource: &ConnectUser{},
		Lister:   &ConnectUserLister{},
	})
}

type ConnectUserLister struct{}

func (l *ConnectUserLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListUsersPaginator(svc, &connect.ListUsersInput{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, user := range resp.UserSummaryList {
				var tags map[string]string
				if user.Arn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: user.Arn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect user: %s", *user.Arn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectUser{
					svc:        svc,
					InstanceID: instance.Id,
					UserID:     user.Id,
					Username:   user.Username,
					ARN:        user.Arn,
					Tags:       tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectUser struct {
	svc        *connect.Client
	InstanceID *string
	UserID     *string
	Username   *string
	ARN        *string
	Tags       map[string]string
}

func (r *ConnectUser) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteUser(ctx, &connect.DeleteUserInput{
		InstanceId: r.InstanceID,
		UserId:     r.UserID,
	})
	return err
}

func (r *ConnectUser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectUser) String() string {
	return *r.Username
}
