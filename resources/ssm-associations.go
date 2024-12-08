package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMAssociationResource = "SSMAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSMAssociationResource,
		Scope:    nuke.Account,
		Resource: &SSMAssociation{},
		Lister:   &SSMAssociationLister{},
	})
}

type SSMAssociationLister struct{}

func (l *SSMAssociationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ssm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ssm.ListAssociationsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListAssociations(params)
		if err != nil {
			return nil, err
		}

		for _, association := range output.Associations {
			resources = append(resources, &SSMAssociation{
				svc:           svc,
				associationID: association.AssociationId,
				instanceID:    association.InstanceId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type SSMAssociation struct {
	svc           *ssm.SSM
	associationID *string
	instanceID    *string
}

func (f *SSMAssociation) Remove(_ context.Context) error {
	_, err := f.svc.DeleteAssociation(&ssm.DeleteAssociationInput{
		AssociationId: f.associationID,
	})

	return err
}

func (f *SSMAssociation) String() string {
	return *f.associationID
}
