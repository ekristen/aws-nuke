package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type ElasticacheUser struct {
	svc      *elasticache.ElastiCache
	userID   *string
	userName *string
}

const ElasticacheUserResource = "ElasticacheUser"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticacheUserResource,
		Scope:  nuke.Account,
		Lister: &ElasticacheUserLister{},
	})
}

type ElasticacheUserLister struct{}

func (l *ElasticacheUserLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticache.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var nextToken *string

	for {
		params := &elasticache.DescribeUsersInput{
			MaxRecords: aws.Int64(100),
			Marker:     nextToken,
		}
		resp, err := svc.DescribeUsers(params)
		if err != nil {
			return nil, err
		}

		for _, user := range resp.Users {
			resources = append(resources, &ElasticacheUser{
				svc:      svc,
				userID:   user.UserId,
				userName: user.UserName,
			})
		}

		// Check if there are more results
		if resp.Marker == nil {
			break // No more results, exit the loop
		}

		// Set the nextToken for the next iteration
		nextToken = resp.Marker
	}

	return resources, nil
}

func (i *ElasticacheUser) Filter() error {
	if ptr.ToString(i.userID) == "default" {
		return fmt.Errorf("cannot delete default user")
	}
	return nil
}

func (i *ElasticacheUser) Remove(_ context.Context) error {
	params := &elasticache.DeleteUserInput{
		UserId: i.userID,
	}

	_, err := i.svc.DeleteUser(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *ElasticacheUser) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", i.userID)
	properties.Set("UserName", i.userName)
	return properties
}

func (i *ElasticacheUser) String() string {
	return *i.userID
}
