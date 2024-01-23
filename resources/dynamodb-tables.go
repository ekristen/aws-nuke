package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const DynamoDBTableResource = "DynamoDBTable"

func init() {
	resource.Register(&resource.Registration{
		Name:   DynamoDBTableResource,
		Scope:  nuke.Account,
		Lister: &DynamoDBTableLister{},
		DependsOn: []string{
			DynamoDBTableItemResource,
		},
	})
}

type DynamoDBTableLister struct{}

func (l *DynamoDBTableLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := dynamodb.New(opts.Session)

	resp, err := svc.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, tableName := range resp.TableNames {
		tags, err := GetTableTags(svc, tableName)

		if err != nil {
			continue
		}

		resources = append(resources, &DynamoDBTable{
			svc:  svc,
			id:   *tableName,
			tags: tags,
		})
	}

	return resources, nil
}

type DynamoDBTable struct {
	svc  *dynamodb.DynamoDB
	id   string
	tags []*dynamodb.Tag
}

func (i *DynamoDBTable) Remove(_ context.Context) error {
	params := &dynamodb.DeleteTableInput{
		TableName: aws.String(i.id),
	}

	_, err := i.svc.DeleteTable(params)
	if err != nil {
		return err
	}

	return nil
}

func GetTableTags(svc *dynamodb.DynamoDB, tableName *string) ([]*dynamodb.Tag, error) {
	result, err := svc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(*tableName),
	})

	if err != nil {
		return make([]*dynamodb.Tag, 0), err
	}

	tags, err := svc.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
		ResourceArn: result.Table.TableArn,
	})

	if err != nil {
		return make([]*dynamodb.Tag, 0), err
	}

	return tags.Tags, nil
}

func (i *DynamoDBTable) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Identifier", i.id)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (i *DynamoDBTable) String() string {
	return i.id
}
