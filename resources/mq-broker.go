package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"        //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/mq" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MQBrokerResource = "MQBroker"

func init() {
	registry.Register(&registry.Registration{
		Name:     MQBrokerResource,
		Scope:    nuke.Account,
		Resource: &MQBroker{},
		Lister:   &MQBrokerLister{},
	})
}

type MQBrokerLister struct{}

func (l *MQBrokerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mq.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mq.ListBrokersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListBrokers(params)
		if err != nil {
			return nil, err
		}

		for _, broker := range resp.BrokerSummaries {
			resources = append(resources, &MQBroker{
				svc:      svc,
				brokerID: broker.BrokerId,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type MQBroker struct {
	svc      *mq.MQ
	brokerID *string
}

func (f *MQBroker) Remove(_ context.Context) error {
	_, err := f.svc.DeleteBroker(&mq.DeleteBrokerInput{
		BrokerId: f.brokerID,
	})

	return err
}

func (f *MQBroker) String() string {
	return *f.brokerID
}
