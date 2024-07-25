package resources

import (
	"context"
	"fmt"

	"strings"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DynamoDBTableItemResource = "DynamoDBTableItem"

func init() {
	registry.Register(&registry.Registration{
		Name:   DynamoDBTableItemResource,
		Scope:  nuke.Account,
		Lister: &DynamoDBTableItemLister{},
	})
}

type DynamoDBTableItemLister struct{}

func (l *DynamoDBTableItemLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := dynamodb.New(opts.Session)

	tableLister := &DynamoDBTableLister{}
	tables, tablesErr := tableLister.List(ctx, o)
	if tablesErr != nil {
		return nil, tablesErr
	}

	resources := make([]resource.Resource, 0)
	for _, dynamoTableResource := range tables {
		dynamoTable, ok := dynamoTableResource.(*DynamoDBTable)
		if !ok {
			// This should never happen (tm).
			logrus.Errorf("Unable to cast DynamoDBTable.")
			continue
		}

		describeParams := &dynamodb.DescribeTableInput{
			TableName: dynamoTable.Name,
		}

		descResp, descErr := svc.DescribeTable(describeParams)
		if descErr != nil {
			return nil, descErr
		}

		keyName := descResp.Table.KeySchema[0].AttributeName
		params := &dynamodb.ScanInput{
			TableName:            dynamoTable.Name,
			ProjectionExpression: aws.String("#key"),
			ExpressionAttributeNames: map[string]*string{
				"#key": keyName,
			},
		}

		scanResp, scanErr := svc.Scan(params)
		if scanErr != nil {
			return nil, scanErr
		}

		for _, itemMap := range scanResp.Items {
			var keyValue string

			for _, value := range itemMap {
				trimmedValue := strings.TrimSpace(value.String())
				keyValue = string([]rune(trimmedValue)[8:(len([]rune(trimmedValue)) - 3)])
			}

			resources = append(resources, &DynamoDBTableItem{
				svc:      svc,
				id:       itemMap,
				table:    dynamoTable,
				keyName:  aws.StringValue(keyName),
				keyValue: keyValue,
			})
		}
	}

	return resources, nil
}

type DynamoDBTableItem struct {
	svc      *dynamodb.DynamoDB
	id       map[string]*dynamodb.AttributeValue
	table    *DynamoDBTable
	keyName  string
	keyValue string
}

func (i *DynamoDBTableItem) Remove(_ context.Context) error {
	params := &dynamodb.DeleteItemInput{
		Key:       i.id,
		TableName: i.table.Name,
	}

	_, err := i.svc.DeleteItem(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *DynamoDBTableItem) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Table", i.table)
	properties.Set("KeyName", i.keyName)
	properties.Set("KeyValue", i.keyValue)
	return properties
}

func (i *DynamoDBTableItem) String() string {
	return fmt.Sprintf("%s -> %s", i.table.String(), i.keyValue)
}
