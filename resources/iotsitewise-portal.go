package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotsitewise"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTSiteWisePortalResource = "IoTSiteWisePortal"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTSiteWisePortalResource,
		Scope:    nuke.Account,
		Resource: &IoTSiteWisePortal{},
		Lister:   &IoTSiteWisePortalLister{},
		DependsOn: []string{
			IoTSiteWiseProjectResource,
			IoTSiteWiseAccessPolicyResource,
		},
	})
}

type IoTSiteWisePortalLister struct{}

func (l *IoTSiteWisePortalLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iotsitewise.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iotsitewise.ListPortalsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListPortals(params)
		if err != nil {
			return nil, err
		}
		for _, item := range resp.PortalSummaries {
			resources = append(resources, &IoTSiteWisePortal{
				svc:  svc,
				ID:   item.Id,
				Name: item.Name,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type IoTSiteWisePortal struct {
	svc  *iotsitewise.IoTSiteWise
	ID   *string
	Name *string
}

func (r *IoTSiteWisePortal) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *IoTSiteWisePortal) Remove(_ context.Context) error {
	_, err := r.svc.DeletePortal(&iotsitewise.DeletePortalInput{
		PortalId: r.ID,
	})

	return err
}

func (r *IoTSiteWisePortal) String() string {
	return *r.ID
}
