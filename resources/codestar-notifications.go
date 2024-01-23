package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CodeStarNotificationRuleResource = "CodeStarNotificationRule"

func init() {
	resource.Register(&resource.Registration{
		Name:   CodeStarNotificationRuleResource,
		Scope:  nuke.Account,
		Lister: &CodeStarNotificationRuleLister{},
	})
}

type CodeStarNotificationRuleLister struct{}

func (l *CodeStarNotificationRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := codestarnotifications.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &codestarnotifications.ListNotificationRulesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListNotificationRules(params)
		if err != nil {
			return nil, err
		}

		for _, notification := range output.NotificationRules {
			descOutput, err := svc.DescribeNotificationRule(&codestarnotifications.DescribeNotificationRuleInput{
				Arn: notification.Arn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &CodeStarNotificationRule{
				svc:  svc,
				id:   notification.Id,
				name: descOutput.Name,
				arn:  notification.Arn,
				tags: descOutput.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CodeStarNotificationRule struct {
	svc  *codestarnotifications.CodeStarNotifications
	id   *string
	name *string
	arn  *string
	tags map[string]*string
}

func (cn *CodeStarNotificationRule) Remove(_ context.Context) error {
	_, err := cn.svc.DeleteNotificationRule(&codestarnotifications.DeleteNotificationRuleInput{
		Arn: cn.arn,
	})

	return err
}

func (cn *CodeStarNotificationRule) String() string {
	return fmt.Sprintf("%s (%s)", *cn.id, *cn.name)
}

func (cn *CodeStarNotificationRule) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range cn.tags {
		properties.SetTag(&key, tag)
	}
	properties.
		Set("Name", cn.name).
		Set("ID", cn.id)
	return properties
}
