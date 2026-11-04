package resources

// DataZone Subscription Targets define where subscribed data gets made
// available (e.g. a Redshift or Glue environment). They are scoped to
// an environment within a domain, so we must list environments first.

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

const DataZoneSubscriptionTargetResource = "DataZoneSubscriptionTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:     DataZoneSubscriptionTargetResource,
		Scope:    nuke.Account,
		Resource: &DataZoneSubscriptionTarget{},
		Lister:   &DataZoneSubscriptionTargetLister{},
		DependsOn: []string{
			DataZoneSubscriptionGrantResource,
		},
	})
}

type DataZoneSubscriptionTargetLister struct {
	mockSvc datazoneiface.DataZoneAPI
}

func (l *DataZoneSubscriptionTargetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc datazoneiface.DataZoneAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = datazone.New(opts.Session)
	}

	resources := make([]resource.Resource, 0)

	projects, err := listDataZoneProjects(svc)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		// Both environments and subscription targets are scoped by project/environment.
		envParams := &datazone.ListEnvironmentsInput{
			DomainIdentifier:  project.DomainID,
			ProjectIdentifier: project.ProjectID,
			MaxResults:        aws.Int64(25),
		}

		for {
			envResp, err := svc.ListEnvironments(envParams)
			if err != nil {
				return nil, err
			}

			for _, env := range envResp.Items {
				// For each environment, list all subscription targets.
				targetParams := &datazone.ListSubscriptionTargetsInput{
					DomainIdentifier:      project.DomainID,
					EnvironmentIdentifier: env.Id,
					MaxResults:            aws.Int64(25),
				}

				for {
					targetResp, err := svc.ListSubscriptionTargets(targetParams)
					if err != nil {
						return nil, err
					}

					for _, target := range targetResp.Items {
						resources = append(resources, &DataZoneSubscriptionTarget{
							svc:           svc,
							DomainID:      project.DomainID,
							EnvironmentID: env.Id,
							ID:            target.Id,
							Name:          target.Name,
						})
					}

					if targetResp.NextToken == nil {
						break
					}
					targetParams.NextToken = targetResp.NextToken
				}
			}

			if envResp.NextToken == nil {
				break
			}
			envParams.NextToken = envResp.NextToken
		}
	}

	return resources, nil
}

type DataZoneSubscriptionTarget struct {
	svc           datazoneiface.DataZoneAPI
	DomainID      *string
	EnvironmentID *string
	ID            *string
	Name          *string
}

func (r *DataZoneSubscriptionTarget) Remove(_ context.Context) error {
	_, err := r.svc.DeleteSubscriptionTarget(&datazone.DeleteSubscriptionTargetInput{
		DomainIdentifier:      r.DomainID,
		EnvironmentIdentifier: r.EnvironmentID,
		Identifier:            r.ID,
	})
	return err
}

func (r *DataZoneSubscriptionTarget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DataZoneSubscriptionTarget) String() string {
	return aws.StringValue(r.ID)
}
