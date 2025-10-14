package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iot" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTPolicyResource = "IoTPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTPolicyResource,
		Scope:    nuke.Account,
		Resource: &IoTPolicy{},
		Lister:   &IoTPolicyLister{},
	})
}

type IoTPolicyLister struct{}

func (l *IoTPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListPoliciesInput{
		PageSize: aws.Int64(25),
	}
	for {
		output, err := svc.ListPolicies(params)
		if err != nil {
			return nil, err
		}

		for _, policy := range output.Policies {
			p := &IoTPolicy{
				svc:  svc,
				name: policy.PolicyName,
			}

			p, err = listIoTPolicyTargets(p)
			if err != nil {
				return nil, err
			}

			p, err = listIoTPolicyDeprecatedVersions(p)
			if err != nil {
				return nil, err
			}

			resources = append(resources, p)
		}
		if output.NextMarker == nil {
			break
		}

		params.Marker = output.NextMarker
	}

	return resources, nil
}

func listIoTPolicyTargets(f *IoTPolicy) (*IoTPolicy, error) {
	var targets []*string

	params := &iot.ListTargetsForPolicyInput{
		PolicyName: f.name,
		PageSize:   aws.Int64(25),
	}

	for {
		output, err := f.svc.ListTargetsForPolicy(params)
		if err != nil {
			return nil, err
		}

		targets = append(targets, output.Targets...)

		if output.NextMarker == nil {
			break
		}

		params.Marker = output.NextMarker
	}

	f.targets = targets

	return f, nil
}

func listIoTPolicyDeprecatedVersions(f *IoTPolicy) (*IoTPolicy, error) {
	var deprecatedVersions []*string

	params := &iot.ListPolicyVersionsInput{
		PolicyName: f.name,
	}

	output, err := f.svc.ListPolicyVersions(params)
	if err != nil {
		return nil, err
	}

	for _, policyVersion := range output.PolicyVersions {
		if !ptr.ToBool(policyVersion.IsDefaultVersion) {
			deprecatedVersions = append(deprecatedVersions, policyVersion.VersionId)
		}
	}

	f.deprecatedVersions = deprecatedVersions

	return f, nil
}

type IoTPolicy struct {
	svc                *iot.IoT
	name               *string
	targets            []*string
	deprecatedVersions []*string
}

func (f *IoTPolicy) Remove(_ context.Context) error {
	// detach attached targets first
	for _, target := range f.targets {
		_, err := f.svc.DetachPolicy(&iot.DetachPolicyInput{
			PolicyName: f.name,
			Target:     target,
		})
		if err != nil {
			return err
		}
	}

	// delete deprecated versions
	for _, version := range f.deprecatedVersions {
		_, err := f.svc.DeletePolicyVersion(&iot.DeletePolicyVersionInput{
			PolicyName:      f.name,
			PolicyVersionId: version,
		})
		if err != nil {
			return err
		}
	}

	_, err := f.svc.DeletePolicy(&iot.DeletePolicyInput{
		PolicyName: f.name,
	})

	return err
}

func (f *IoTPolicy) String() string {
	return *f.name
}
