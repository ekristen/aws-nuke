package resources

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DynamoDBTableResource = "DynamoDBTable"

func init() {
	registry.Register(&registry.Registration{
		Name:   DynamoDBTableResource,
		Scope:  nuke.Account,
		Lister: &DynamoDBTableLister{},
		Settings: []string{
			"DisableDeletionProtection",
		},
		DependsOn: []string{
			DynamoDBTableItemResource,
		},
	})
}

type DynamoDBTableLister struct {
	mockSvc dynamodbiface.DynamoDBAPI
}

func (l *DynamoDBTableLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc dynamodbiface.DynamoDBAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = dynamodb.New(opts.Session)
	}

	resp, err := svc.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}

	for _, tableName := range resp.TableNames {
		table, err := svc.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: tableName,
		})
		if err != nil {
			logrus.WithError(err).Warn("unable to describe table")
			continue
		}

		tags, err := svc.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
			ResourceArn: table.Table.TableArn,
		})
		if err != nil {
			logrus.WithError(err).Warn("unable to list tags of resource")
			continue
		}

		resources = append(resources, &DynamoDBTable{
			svc:        svc,
			id:         tableName,
			protection: table.Table.DeletionProtectionEnabled,
			Name:       tableName,
			Tags:       tags.Tags,
		})
	}

	return resources, nil
}

type DynamoDBTable struct {
	svc        dynamodbiface.DynamoDBAPI
	settings   *settings.Setting
	id         *string `property:"Identifier"` // TODO(v4): remove this
	protection *bool
	Name       *string
	Tags       []*dynamodb.Tag
}

func (r *DynamoDBTable) Remove(_ context.Context) error {
	if err := r.DisableDeletionProtection(); err != nil {
		return err
	}

	params := &dynamodb.DeleteTableInput{
		TableName: r.Name,
	}

	_, err := r.svc.DeleteTable(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *DynamoDBTable) DisableDeletionProtection() error {
	if !r.settings.GetBool("DisableDeletionProtection") {
		return nil
	}

	if ptr.ToBool(r.protection) {
		params := &dynamodb.UpdateTableInput{
			TableName:                 r.Name,
			DeletionProtectionEnabled: ptr.Bool(false),
		}

		_, err := r.svc.UpdateTable(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *DynamoDBTable) Settings(setting *settings.Setting) {
	r.settings = setting
}

func (r *DynamoDBTable) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DynamoDBTable) String() string {
	return ptr.ToString(r.Name)
}
