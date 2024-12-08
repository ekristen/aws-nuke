package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ECSCapacityProviderResource = "ECSCapacityProvider"

func init() {
	registry.Register(&registry.Registration{
		Name:     ECSCapacityProviderResource,
		Scope:    nuke.Account,
		Resource: &ECSCapacityProvider{},
		Lister:   &ECSCapacityProviderLister{},
	})
}

type ECSCapacityProvider struct {
	svc    *ecs.ECS
	ARN    *string
	Name   *string
	Status *string
	Tags   []*ecs.Tag
}

func (r *ECSCapacityProvider) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCapacityProvider(&ecs.DeleteCapacityProviderInput{
		CapacityProvider: r.ARN,
	})

	return err
}

func (r *ECSCapacityProvider) Filter() error {
	// The FARGATE and FARGATE_SPOT capacity providers cannot be deleted
	if *r.Name == "FARGATE" || *r.Name == "FARGATE_SPOT" {
		return fmt.Errorf("unable to delete, fargate managed")
	}

	return nil
}

func (r *ECSCapacityProvider) String() string {
	return *r.ARN
}

func (r *ECSCapacityProvider) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

//---------------------------------------

type ECSCapacityProviderLister struct{}

func (l *ECSCapacityProviderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := ecs.New(opts.Session)

	params := &ecs.DescribeCapacityProvidersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeCapacityProviders(params)
		if err != nil {
			return nil, err
		}

		for _, capacityProviders := range output.CapacityProviders {
			resources = append(resources, &ECSCapacityProvider{
				svc:    svc,
				ARN:    capacityProviders.CapacityProviderArn,
				Name:   capacityProviders.Name,
				Status: capacityProviders.Status,
				Tags:   capacityProviders.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}
