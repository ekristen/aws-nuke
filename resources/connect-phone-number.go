package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/connect"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConnectPhoneNumberResource = "ConnectPhoneNumber"

func init() {
	registry.Register(&registry.Registration{
		Name:     ConnectPhoneNumberResource,
		Scope:    nuke.Account,
		Resource: &ConnectPhoneNumber{},
		Lister:   &ConnectPhoneNumberLister{},
	})
}

type ConnectPhoneNumberLister struct{}

func (l *ConnectPhoneNumberLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := connect.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	instances, err := listConnectInstances(ctx, svc)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		paginator := connect.NewListPhoneNumbersV2Paginator(svc, &connect.ListPhoneNumbersV2Input{
			InstanceId: instance.Id,
		})

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, phone := range resp.ListPhoneNumbersSummaryList {
				var tags map[string]string
				if phone.PhoneNumberArn != nil {
					tagsResp, err := svc.ListTagsForResource(ctx, &connect.ListTagsForResourceInput{
						ResourceArn: phone.PhoneNumberArn,
					})
					if err != nil {
						opts.Logger.Warnf("unable to fetch tags for connect phone number: %s", *phone.PhoneNumberArn)
					} else {
						tags = tagsResp.Tags
					}
				}

				resources = append(resources, &ConnectPhoneNumber{
					svc:             svc,
					PhoneNumberID:   phone.PhoneNumberId,
					PhoneNumber:     phone.PhoneNumber,
					PhoneNumberType: string(phone.PhoneNumberType),
					CountryCode:     string(phone.PhoneNumberCountryCode),
					ARN:             phone.PhoneNumberArn,
					InstanceID:      phone.InstanceId,
					Tags:            tags,
				})
			}
		}
	}

	return resources, nil
}

type ConnectPhoneNumber struct {
	svc             *connect.Client
	PhoneNumberID   *string
	PhoneNumber     *string
	PhoneNumberType string
	CountryCode     string
	ARN             *string
	InstanceID      *string
	Tags            map[string]string
}

func (r *ConnectPhoneNumber) Remove(ctx context.Context) error {
	_, err := r.svc.ReleasePhoneNumber(ctx, &connect.ReleasePhoneNumberInput{
		PhoneNumberId: r.PhoneNumberID,
	})
	return err
}

func (r *ConnectPhoneNumber) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ConnectPhoneNumber) String() string {
	return *r.PhoneNumber
}
