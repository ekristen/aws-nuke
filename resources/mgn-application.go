package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mgn"
	"github.com/aws/aws-sdk-go-v2/service/mgn/types"
	"github.com/aws/smithy-go"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	MGNApplicationResource                      = "MGNApplication"
	mgnApplicationUninitializedAccountException = "UninitializedAccountException"
)

func init() {
	registry.Register(&registry.Registration{
		Name:     MGNApplicationResource,
		Scope:    nuke.Account,
		Resource: &MGNApplication{},
		Lister:   &MGNApplicationLister{},
	})
}

type MGNApplicationLister struct{}

func (l *MGNApplicationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mgn.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &mgn.ListApplicationsInput{
		MaxResults: aws.Int32(50),
	}

	for {
		output, err := svc.ListApplications(ctx, params)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == mgnApplicationUninitializedAccountException {
				return nil, nil
			}
			return nil, err
		}

		for i := range output.Items {
			application := &output.Items[i]
			mgnApp := &MGNApplication{
				svc:                  svc,
				application:          application,
				ApplicationID:        application.ApplicationID,
				Arn:                  application.Arn,
				Name:                 application.Name,
				Description:          application.Description,
				IsArchived:           application.IsArchived,
				CreationDateTime:     application.CreationDateTime,
				LastModifiedDateTime: application.LastModifiedDateTime,
				Tags:                 application.Tags,
			}
			resources = append(resources, mgnApp)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MGNApplication struct {
	svc         *mgn.Client        `description:"-"`
	application *types.Application `description:"-"`

	// Exposed properties
	ApplicationID        *string           `description:"The unique identifier of the application"`
	Arn                  *string           `description:"The ARN of the application"`
	Name                 *string           `description:"The name of the application"`
	Description          *string           `description:"The description of the application"`
	IsArchived           *bool             `description:"Whether the application is archived"`
	CreationDateTime     *string           `description:"The date and time the application was created"`
	LastModifiedDateTime *string           `description:"The date and time the application was last modified"`
	Tags                 map[string]string `description:"The tags associated with the application"`
}

func (f *MGNApplication) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteApplication(ctx, &mgn.DeleteApplicationInput{
		ApplicationID: f.application.ApplicationID,
	})

	return err
}

func (f *MGNApplication) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(f)
}

func (f *MGNApplication) String() string {
	return *f.ApplicationID
}
