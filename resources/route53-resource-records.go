package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const Route53ResourceRecordSetResource = "Route53ResourceRecordSet"

func init() {
	resource.Register(&resource.Registration{
		Name:   Route53ResourceRecordSetResource,
		Scope:  nuke.Account,
		Lister: &Route53ResourceRecordSetLister{},
	})
}

type Route53ResourceRecordSetLister struct{}

func (l *Route53ResourceRecordSetLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := route53.New(opts.Session)

	resources := make([]resource.Resource, 0)

	zoneLister := &Route53HostedZoneLister{}
	sub, err := zoneLister.List(ctx, o)
	if err != nil {
		return nil, err
	}

	for _, r := range sub {
		zone := r.(*Route53HostedZone)
		rrs, err := ListResourceRecordsForZone(svc, zone.id, zone.name)
		if err != nil {
			return nil, err
		}

		resources = append(resources, rrs...)
	}

	return resources, nil
}

func ListResourceRecordsForZone(svc *route53.Route53, zoneId *string, zoneName *string) ([]resource.Resource, error) {
	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: zoneId,
	}

	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListResourceRecordSets(params)
		if err != nil {
			return nil, err
		}

		for _, rrs := range resp.ResourceRecordSets {
			resources = append(resources, &Route53ResourceRecordSet{
				svc:            svc,
				hostedZoneId:   zoneId,
				hostedZoneName: zoneName,
				data:           rrs,
			})
		}

		// make sure to list all with more than 100 records
		if *resp.IsTruncated {
			params.StartRecordName = resp.NextRecordName
			continue
		}

		break
	}

	return resources, nil
}

type Route53ResourceRecordSet struct {
	svc            *route53.Route53
	hostedZoneId   *string
	hostedZoneName *string
	data           *route53.ResourceRecordSet
	changeId       *string
}

func (r *Route53ResourceRecordSet) Filter() error {
	if *r.data.Type == "NS" && *r.hostedZoneName == *r.data.Name {
		return fmt.Errorf("cannot delete NS record")
	}

	if *r.data.Type == "SOA" {
		return fmt.Errorf("cannot delete SOA record")
	}

	return nil
}

func (r *Route53ResourceRecordSet) Remove(_ context.Context) error {
	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: r.hostedZoneId,
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				&route53.Change{
					Action:            aws.String("DELETE"),
					ResourceRecordSet: r.data,
				},
			},
		},
	}

	resp, err := r.svc.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}

	r.changeId = resp.ChangeInfo.Id

	return nil
}

func (r *Route53ResourceRecordSet) Properties() types.Properties {
	return types.NewProperties().
		Set("Name", r.data.Name).
		Set("Type", r.data.Type)
}

func (r *Route53ResourceRecordSet) String() string {
	return *r.data.Name
}
