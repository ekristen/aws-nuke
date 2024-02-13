package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const S3AccessPointResource = "S3AccessPoint"

func init() {
	registry.Register(&registry.Registration{
		Name:   S3AccessPointResource,
		Scope:  nuke.Account,
		Lister: &S3AccessPointLister{},
	})
}

type S3AccessPointLister struct{}

func (l *S3AccessPointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	stsSvc := sts.New(opts.Session)
	callerID, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	accountId := callerID.Account

	var resources []resource.Resource
	svc := s3control.New(opts.Session)
	for {
		params := &s3control.ListAccessPointsInput{
			AccountId: accountId,
		}

		resp, err := svc.ListAccessPoints(params)
		if err != nil {
			return nil, err
		}

		for _, accessPoint := range resp.AccessPointList {
			resources = append(resources, &S3AccessPoint{
				svc:         svc,
				accountId:   accountId,
				accessPoint: accessPoint,
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
	svc         *s3control.S3Control
	accountId   *string
	accessPoint *s3control.AccessPoint
}

func (e *S3AccessPoint) Remove(_ context.Context) error {
	_, err := e.svc.DeleteAccessPoint(&s3control.DeleteAccessPointInput{
		AccountId: e.accountId,
		Name:      aws.String(*e.accessPoint.Name),
	})
	return err
}

func (e *S3AccessPoint) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("AccessPointArn", e.accessPoint.AccessPointArn).
		Set("Alias", e.accessPoint.Alias).
		Set("Bucket", e.accessPoint.Bucket).
		Set("Name", e.accessPoint.Name).
		Set("NetworkOrigin", e.accessPoint.NetworkOrigin)

	return properties
}

func (e *S3AccessPoint) String() string {
	return ptr.ToString(e.accessPoint.AccessPointArn)
}
