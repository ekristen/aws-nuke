package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"          //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/glue" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type GlueSession struct {
	svc *glue.Glue
	id  *string
}

const GlueSessionResource = "GlueSession"

func init() {
	registry.Register(&registry.Registration{
		Name:     GlueSessionResource,
		Scope:    nuke.Account,
		Resource: &GlueSession{},
		Lister:   &GlueSessionLister{},
	})
}

type GlueSessionLister struct{}

func (l *GlueSessionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.ListSessionsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		output, err := svc.ListSessions(params)
		if err != nil {
			return nil, err
		}

		for _, session := range output.Sessions {
			resources = append(resources, &GlueSession{
				svc: svc,
				id:  session.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *GlueSession) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSession(&glue.DeleteSessionInput{
		Id: f.id,
	})

	return err
}

func (f *GlueSession) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.id)

	return properties
}

func (f *GlueSession) String() string {
	return *f.id
}
