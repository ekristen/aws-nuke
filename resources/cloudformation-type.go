package resources

import (
	"context"

	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/cloudformation" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFormationTypeResource = "CloudFormationType"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFormationTypeResource,
		Scope:    nuke.Account,
		Resource: &CloudFormationType{},
		Lister:   &CloudFormationTypeLister{},
	})
}

type CloudFormationTypeLister struct{}

func (l *CloudFormationTypeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudformation.New(opts.Session)

	params := &cloudformation.ListTypesInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListTypes(params)
		if err != nil {
			return nil, err
		}
		for _, typeSummary := range resp.TypeSummaries {
			resources = append(resources, &CloudFormationType{
				svc:         svc,
				typeSummary: typeSummary,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudFormationType struct {
	svc         cloudformationiface.CloudFormationAPI
	typeSummary *cloudformation.TypeSummary
}

func (cfs *CloudFormationType) findAllVersionSummaries() ([]*cloudformation.TypeVersionSummary, error) {
	typeVersionSummaries := make([]*cloudformation.TypeVersionSummary, 0)
	page := 0
	params := &cloudformation.ListTypeVersionsInput{
		Arn: cfs.typeSummary.TypeArn,
	}
	for {
		logrus.Infof("CloudFormationType loading type versions arn=%s page=%d", *cfs.typeSummary.TypeArn, page)
		resp, err := cfs.svc.ListTypeVersions(params)
		if err != nil {
			return nil, err
		}

		typeVersionSummaries = append(typeVersionSummaries, resp.TypeVersionSummaries...)
		if resp.NextToken == nil {
			return typeVersionSummaries, nil
		}

		params.NextToken = resp.NextToken
		page++
	}
}

func (cfs *CloudFormationType) Remove(_ context.Context) error {
	typeVersionSummaries, loadErr := cfs.findAllVersionSummaries()
	if loadErr != nil {
		return loadErr
	}

	failed := false
	for _, typeVersionSummary := range typeVersionSummaries {
		if *typeVersionSummary.IsDefaultVersion {
			logrus.Infof("CloudFormationType ignoring default version type=%s version=%s",
				*cfs.typeSummary.TypeArn, *typeVersionSummary.VersionId)
		} else {
			logrus.Infof("CloudFormationType removing type=%s version=%s",
				*cfs.typeSummary.TypeArn, *typeVersionSummary.VersionId)
			if _, err := cfs.svc.DeregisterType(&cloudformation.DeregisterTypeInput{
				VersionId: typeVersionSummary.VersionId,
				TypeName:  typeVersionSummary.TypeName,
				Type:      typeVersionSummary.Type,
			}); err != nil {
				logrus.Errorf("CloudFormationType failed removing type=%s version=%s type=%s arn=%s error=%s",
					*cfs.typeSummary.TypeName, *typeVersionSummary.VersionId, *typeVersionSummary.Type,
					*cfs.typeSummary.TypeArn, err.Error())
				failed = true
			}
		}
	}

	if failed {
		return fmt.Errorf("unable to remove all CloudFormationType versions arn=%s", *cfs.typeSummary.TypeArn)
	}

	_, err := cfs.svc.DeregisterType(&cloudformation.DeregisterTypeInput{
		Arn: cfs.typeSummary.TypeArn,
	})

	return err
}

func (cfs *CloudFormationType) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", cfs.typeSummary.TypeName)
	properties.Set("Type", cfs.typeSummary.Type)

	return properties
}

func (cfs *CloudFormationType) String() string {
	return *cfs.typeSummary.TypeArn
}
