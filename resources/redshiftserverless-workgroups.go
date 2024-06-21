package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type RedshiftServerlessWorkgroup struct {
	svc       *redshiftserverless.RedshiftServerless
	workgroup *redshiftserverless.Workgroup
}

const RedshiftServerlessWorkgroupResource = "RedshiftServerlessWorkgroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   RedshiftServerlessWorkgroupResource,
		Scope:  nuke.Account,
		Lister: &RedshiftServerlessWorkgroupLister{},
	})
}

type RedshiftServerlessWorkgroupLister struct{}

func (l *RedshiftServerlessWorkgroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshiftserverless.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshiftserverless.ListWorkgroupsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListWorkgroups(params)
		if err != nil {
			return nil, err
		}

		for _, workgroup := range output.Workgroups {
			resources = append(resources, &RedshiftServerlessWorkgroup{
				svc:       svc,
				workgroup: workgroup,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (w *RedshiftServerlessWorkgroup) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreationDate", w.workgroup.CreationDate).
		Set("Namespace", w.workgroup.NamespaceName).
		Set("WorkgroupName", w.workgroup.WorkgroupName)

	return properties
}

func (w *RedshiftServerlessWorkgroup) Remove(_ context.Context) error {
	_, err := w.svc.DeleteWorkgroup(&redshiftserverless.DeleteWorkgroupInput{
		WorkgroupName: w.workgroup.WorkgroupName,
	})

	return err
}

func (w *RedshiftServerlessWorkgroup) String() string {
	return ptr.ToString(w.workgroup.WorkgroupName)
}
