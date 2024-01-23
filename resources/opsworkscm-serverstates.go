package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/opsworkscm"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const OpsWorksCMServerStateResource = "OpsWorksCMServerState"

func init() {
	resource.Register(&resource.Registration{
		Name:   OpsWorksCMServerStateResource,
		Scope:  nuke.Account,
		Lister: &OpsWorksCMServerStateLister{},
	})
}

type OpsWorksCMServerStateLister struct{}

func (l *OpsWorksCMServerStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworkscm.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &opsworkscm.DescribeServersInput{}

	output, err := svc.DescribeServers(params)
	if err != nil {
		return nil, err
	}

	for _, server := range output.Servers {
		resources = append(resources, &OpsWorksCMServerState{
			svc:    svc,
			name:   server.ServerName,
			status: server.Status,
		})
	}

	return resources, nil
}

type OpsWorksCMServerState struct {
	svc    *opsworkscm.OpsWorksCM
	name   *string
	status *string
}

func (f *OpsWorksCMServerState) Remove(_ context.Context) error {
	return nil
}

func (f *OpsWorksCMServerState) String() string {
	return *f.name
}

func (f *OpsWorksCMServerState) Filter() error {
	if *f.status == "CREATING" {
		return nil
	} else {
		return fmt.Errorf("available for transition")
	}
}
