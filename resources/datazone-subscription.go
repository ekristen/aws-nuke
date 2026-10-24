package resources

// DataZone Subscriptions represent approved access requests to data assets.
// They must be cancelled before the domain can be deleted.

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datazone"
	"github.com/aws/aws-sdk-go/service/datazone/datazoneiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DataZoneSubscriptionResource = "DataZoneSubscription"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneSubscriptionResource,
		Scope:    nuke.Account,
		Resource: &DataZoneSubscription{},
		Lister:   &DataZoneSubscriptionLister{},
		DependsOn: []string{
			DataZoneSubscriptionGrantResource,
			DataZoneSubscriptionTargetResource,
		},
	})
}

type DataZoneSubscriptionLister struct {
	mockSvc datazoneiface.DataZoneAPI
}

func (l *DataZoneSubscriptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc datazoneiface.DataZoneAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = datazone.New(opts.Session)
	}

	subscriptions, err := listDataZoneSubscriptions(svc)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0, len(subscriptions))
	for _, sub := range subscriptions {
		resources = append(resources, &DataZoneSubscription{
			svc:      svc,
			DomainID: sub.DomainID,
			ID:       sub.ID,
			Status:   sub.Status,
		})
	}

	return resources, nil
}

type dataZoneSubscriptionRef struct {
	DomainID *string
	ID       *string
	Status   *string
}

// ListSubscriptions requires a project or listing filter. Search by every
// project as both owner and approver, then deduplicate subscriptions returned
// by both queries.
func listDataZoneSubscriptions(svc datazoneiface.DataZoneAPI) ([]dataZoneSubscriptionRef, error) {
	projects, err := listDataZoneProjects(svc)
	if err != nil {
		return nil, err
	}

	subscriptions := make([]dataZoneSubscriptionRef, 0)
	seen := make(map[string]struct{})
	for _, project := range projects {
		for _, asApprover := range []bool{false, true} {
			params := &datazone.ListSubscriptionsInput{
				DomainIdentifier: project.DomainID,
				MaxResults:       aws.Int64(25),
			}
			if asApprover {
				params.ApproverProjectId = project.ProjectID
			} else {
				params.OwningProjectId = project.ProjectID
			}

			for {
				resp, err := svc.ListSubscriptions(params)
				if err != nil {
					return nil, err
				}
				for _, sub := range resp.Items {
					key := aws.StringValue(project.DomainID) + "/" + aws.StringValue(sub.Id)
					if _, ok := seen[key]; ok {
						continue
					}
					seen[key] = struct{}{}
					subscriptions = append(subscriptions, dataZoneSubscriptionRef{
						DomainID: project.DomainID,
						ID:       sub.Id,
						Status:   aws.String(aws.StringValue(sub.Status)),
					})
				}
				if resp.NextToken == nil {
					break
				}
				params.NextToken = resp.NextToken
			}
		}
	}

	return subscriptions, nil
}

type DataZoneSubscription struct {
	svc      datazoneiface.DataZoneAPI
	DomainID *string
	ID       *string
	Status   *string
}

// Remove cancels the subscription. DataZone uses CancelSubscription
// rather than DeleteSubscription for active subscriptions.
func (r *DataZoneSubscription) Remove(_ context.Context) error {
	_, err := r.svc.CancelSubscription(&datazone.CancelSubscriptionInput{
		DomainIdentifier: r.DomainID,
		Identifier:       r.ID,
	})
	return err
}

func (r *DataZoneSubscription) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DataZoneSubscription) String() string {
	return aws.StringValue(r.ID)
}
