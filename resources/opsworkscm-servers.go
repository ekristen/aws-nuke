package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opsworkscm"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const OpsWorksCMServerResource = "OpsWorksCMServer"

func init() {
	registry.Register(&registry.Registration{
		Name:   OpsWorksCMServerResource,
		Scope:  nuke.Account,
		Lister: &OpsWorksCMServerLister{},
	})
}

type OpsWorksCMServerLister struct{}

func (l *OpsWorksCMServerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworkscm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &opsworkscm.DescribeServersInput{}

	output, err := svc.DescribeServers(params)
	if err != nil {
		return nil, err
	}

	for _, server := range output.Servers {
		resources = append(resources, &OpsWorksCMServer{
			svc:  svc,
			name: server.ServerName,
		})
	}

	return resources, nil
}

type OpsWorksCMServer struct {
	svc    *opsworkscm.OpsWorksCM
	name   *string
	status *string
}

func (f *OpsWorksCMServer) Remove(_ context.Context) error {
	_, err := f.svc.DeleteServer(&opsworkscm.DeleteServerInput{
		ServerName: f.name,
	})

	return err
}

func (f *OpsWorksCMServer) String() string {
	return *f.name
}
