package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SNSTopicResource = "SNSTopic"

func init() {
	registry.Register(&registry.Registration{
		Name:   SNSTopicResource,
		Scope:  nuke.Account,
		Lister: &SNSTopicLister{},
	})
}

type SNSTopicLister struct{}

func (l *SNSTopicLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sns.New(opts.Session)

	topics := make([]*sns.Topic, 0)

	params := &sns.ListTopicsInput{}

	err := svc.ListTopicsPages(params, func(page *sns.ListTopicsOutput, lastPage bool) bool {
		topics = append(topics, page.Topics...)
		return true
	})
	if err != nil {
		return nil, err
	}
	resources := make([]resource.Resource, 0)
	for _, topic := range topics {
		tags, err := svc.ListTagsForResource(&sns.ListTagsForResourceInput{
			ResourceArn: topic.TopicArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &SNSTopic{
			svc:  svc,
			id:   topic.TopicArn,
			tags: tags.Tags,
		})
	}
	return resources, nil
}

type SNSTopic struct {
	svc  *sns.SNS
	id   *string
	tags []*sns.Tag
}

func (topic *SNSTopic) Remove(_ context.Context) error {
	_, err := topic.svc.DeleteTopic(&sns.DeleteTopicInput{
		TopicArn: topic.id,
	})
	return err
}

func (topic *SNSTopic) Properties() types.Properties {
	properties := types.NewProperties()

	for _, tag := range topic.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.Set("TopicARN", topic.id)

	return properties
}

func (topic *SNSTopic) String() string {
	return fmt.Sprintf("TopicARN: %s", *topic.id)
}
