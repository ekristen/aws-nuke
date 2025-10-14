package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/iot" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTAuthorizerResource = "IoTAuthorizer"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTAuthorizerResource,
		Scope:    nuke.Account,
		Resource: &IoTAuthorizer{},
		Lister:   &IoTAuthorizerLister{},
	})
}

type IoTAuthorizerLister struct{}

func (l *IoTAuthorizerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListAuthorizersInput{}

	output, err := svc.ListAuthorizers(params)
	if err != nil {
		return nil, err
	}

	for _, authorizer := range output.Authorizers {
		resources = append(resources, &IoTAuthorizer{
			svc:  svc,
			name: authorizer.AuthorizerName,
		})
	}

	return resources, nil
}

type IoTAuthorizer struct {
	svc  *iot.IoT
	name *string
}

func (f *IoTAuthorizer) Remove(_ context.Context) error {
	if _, err := f.svc.UpdateAuthorizer(&iot.UpdateAuthorizerInput{
		AuthorizerName: f.name,
		Status:         ptr.String(iot.AuthorizerStatusInactive),
	}); err != nil {
		return err
	}

	_, err := f.svc.DeleteAuthorizer(&iot.DeleteAuthorizerInput{
		AuthorizerName: f.name,
	})

	return err
}

func (f *IoTAuthorizer) String() string {
	return *f.name
}
