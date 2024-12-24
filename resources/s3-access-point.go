package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/s3control"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3AccessPointResource = "S3AccessPoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3AccessPointResource,
		Scope:    nuke.Account,
		Resource: &S3AccessPoint{},
		Lister:   &S3AccessPointLister{},
	})
}

type S3AccessPointLister struct{}

func (l *S3AccessPointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := s3control.New(opts.Session)

	for {
		params := &s3control.ListAccessPointsInput{
			AccountId: opts.AccountID,
		}

		resp, err := svc.ListAccessPoints(params)
		if err != nil {
			return nil, err
		}

		for _, accessPoint := range resp.AccessPointList {
			resources = append(resources, &S3AccessPoint{
				svc:           svc,
				accountID:     opts.AccountID,
				Name:          accessPoint.Name,
				ARN:           accessPoint.AccessPointArn,
				Alias:         accessPoint.Alias,
				Bucket:        accessPoint.Bucket,
				NetworkOrigin: accessPoint.NetworkOrigin,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type S3AccessPoint struct {
	svc           *s3control.S3Control
	accountID     *string
	Name          *string
	ARN           *string
	Alias         *string
	Bucket        *string
	NetworkOrigin *string
}

func (r *S3AccessPoint) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAccessPoint(&s3control.DeleteAccessPointInput{
		AccountId: r.accountID,
		Name:      r.Name,
	})
	return err
}

func (r *S3AccessPoint) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r).
		Set("AccessPointArn", r.ARN) // TODO(ek): this is an alias, should be deprecated for ARN
}

func (r *S3AccessPoint) String() string {
	return ptr.ToString(r.ARN) // TODO(ek): this should be the Name not the ARN
}
