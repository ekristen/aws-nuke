package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"

	liberror "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EMRServerlessApplicationResource = "EMRServerlessApplication"

func init() {
	registry.Register(&registry.Registration{
		Name:     EMRServerlessApplicationResource,
		Scope:    nuke.Account,
		Resource: &EMRServerlessApplication{},
		Lister:   &EMRServerlessApplicationLister{},
		DependsOn: []string{
			EMRServerlessJobRunResource,
		},
	})
}

type EMRServerlessApplicationLister struct{}

func (l *EMRServerlessApplicationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := emrserverless.NewFromConfig(*opts.Config)

	var resources []resource.Resource

	params := &emrserverless.ListApplicationsInput{
		MaxResults: aws.Int32(50),
	}

	paginator := emrserverless.NewListApplicationsPaginator(svc, params)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, app := range page.Applications {
			descOutput, err := svc.GetApplication(ctx, &emrserverless.GetApplicationInput{
				ApplicationId: app.Id,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &EMRServerlessApplication{
				svc:          svc,
				ID:           app.Id,
				Name:         app.Name,
				Type:         app.Type,
				State:        app.State,
				ARN:          app.Arn,
				CreatedAt:    app.CreatedAt,
				UpdatedAt:    app.UpdatedAt,
				Tags:         descOutput.Application.Tags,
				ReleaseLabel: descOutput.Application.ReleaseLabel,
				Architecture: descOutput.Application.Architecture,
			})
		}
	}

	return resources, nil
}

type EMRServerlessApplication struct {
	svc          *emrserverless.Client
	ID           *string
	Name         *string
	Type         *string
	State        types.ApplicationState
	ARN          *string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Tags         map[string]string
	ReleaseLabel *string
	Architecture types.Architecture
}

func (r *EMRServerlessApplication) Remove(ctx context.Context) error {
	// Get current state
	appOutput, err := r.svc.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: r.ID,
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}

	r.State = appOutput.Application.State

	if r.State == types.ApplicationStateStarted || r.State == types.ApplicationStateStarting {
		_, err := r.svc.StopApplication(ctx, &emrserverless.StopApplicationInput{
			ApplicationId: r.ID,
		})
		if err != nil && !strings.Contains(err.Error(), "Application is not in a valid state") {
			return fmt.Errorf("failed to stop application: %w", err)
		}
		return liberror.ErrHoldResource("waiting for application to stop")
	}

	// If still in transitional state, wait
	if r.State == types.ApplicationStateStopping || r.State == types.ApplicationStateCreating {
		return liberror.ErrHoldResource("waiting for application state transition")
	}

	_, err = r.svc.DeleteApplication(ctx, &emrserverless.DeleteApplicationInput{
		ApplicationId: r.ID,
	})
	return err
}

func (r *EMRServerlessApplication) HandleWait(ctx context.Context) error {
	var notFound *types.ResourceNotFoundException
	appOutput, err := r.svc.GetApplication(ctx, &emrserverless.GetApplicationInput{
		ApplicationId: r.ID,
	})
	if err != nil {
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}

	r.State = appOutput.Application.State

	if r.State == types.ApplicationStateTerminated {
		return nil
	}

	return liberror.ErrWaitResource("waiting for application deletion to complete")
}

func (r *EMRServerlessApplication) Filter() error {
	if r.State == types.ApplicationStateTerminated {
		return fmt.Errorf("already terminated")
	}
	return nil
}

func (r *EMRServerlessApplication) Properties() libtypes.Properties {
	properties := libtypes.NewProperties()
	properties.
		Set("ID", r.ID).
		Set("Name", r.Name).
		Set("Type", r.Type).
		Set("State", string(r.State)).
		Set("ARN", r.ARN).
		Set("ReleaseLabel", r.ReleaseLabel).
		Set("Architecture", string(r.Architecture))

	if r.CreatedAt != nil {
		properties.Set("CreatedAt", r.CreatedAt.Format(time.RFC3339))
	}
	if r.UpdatedAt != nil {
		properties.Set("UpdatedAt", r.UpdatedAt.Format(time.RFC3339))
	}

	for key, val := range r.Tags {
		properties.SetTag(&key, &val)
	}

	return properties
}

func (r *EMRServerlessApplication) String() string {
	return *r.ID
}
