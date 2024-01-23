package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTAuthorizerResource = "IoTAuthorizer"

func init() {
	resource.Register(&resource.Registration{
		Name:   IoTAuthorizerResource,
		Scope:  nuke.Account,
		Lister: &IoTAuthorizerLister{},
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
	_, err := f.svc.DeleteAuthorizer(&iot.DeleteAuthorizerInput{
		AuthorizerName: f.name,
	})

	return err
}

func (f *IoTAuthorizer) String() string {
	return *f.name
}
