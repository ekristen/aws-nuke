package resources

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/awserr"  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IAMOpenIDConnectProviderResource = "IAMOpenIDConnectProvider"

func init() {
	registry.Register(&registry.Registration{
		Name:     IAMOpenIDConnectProviderResource,
		Scope:    nuke.Account,
		Resource: &IAMOpenIDConnectProvider{},
		Lister:   &IAMOpenIDConnectProviderLister{},
	})
}

type IAMOpenIDConnectProviderLister struct{}

func (l *IAMOpenIDConnectProviderLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iam.New(opts.Session)

	listParams := &iam.ListOpenIDConnectProvidersInput{}
	resources := make([]resource.Resource, 0)

	resp, err := svc.ListOpenIDConnectProviders(listParams)
	if err != nil {
		return nil, err
	}

	var inaccessibleOpenIDConnectProvider bool

	for _, out := range resp.OpenIDConnectProviderList {
		params := &iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: out.Arn,
		}
		resp, err := svc.GetOpenIDConnectProvider(params)

		if err != nil {
			var awsError awserr.Error
			if errors.As(err, &awsError) {
				if awsError.Code() == "AccessDenied" {
					inaccessibleOpenIDConnectProvider = true
					logrus.WithError(err).WithField("arn", out.Arn).Debug("inaccessible openIDConnectProvider")
					continue
				} else {
					logrus.WithError(err).WithField("arn", out.Arn).Error("unable to list openIDConnectProvider")
				}
			}
		}

		resources = append(resources, &IAMOpenIDConnectProvider{
			svc:  svc,
			arn:  *out.Arn,
			tags: resp.Tags,
		})
	}

	if inaccessibleOpenIDConnectProvider {
		logrus.Warn("one or more OpenIDConnectProviders were inaccessible, debug logging will contain more information")
	}

	return resources, nil
}

type IAMOpenIDConnectProvider struct {
	svc  iamiface.IAMAPI
	arn  string
	tags []*iam.Tag
}

func (e *IAMOpenIDConnectProvider) Remove(_ context.Context) error {
	_, err := e.svc.DeleteOpenIDConnectProvider(&iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMOpenIDConnectProvider) String() string {
	return e.arn
}

func (e *IAMOpenIDConnectProvider) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Arn", e.arn)

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
