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

type RedshiftServerlessNamespace struct {
	svc       *redshiftserverless.RedshiftServerless
	namespace *redshiftserverless.Namespace
}

const RedshiftServerlessNamespaceResource = "RedshiftServerlessNamespace"

func init() {
	registry.Register(&registry.Registration{
		Name:     RedshiftServerlessNamespaceResource,
		Scope:    nuke.Account,
		Resource: &RedshiftServerlessNamespace{},
		Lister:   &RedshiftServerlessNamespaceLister{},
	})
}

type RedshiftServerlessNamespaceLister struct{}

func (l *RedshiftServerlessNamespaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := redshiftserverless.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &redshiftserverless.ListNamespacesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListNamespaces(params)
		if err != nil {
			return nil, err
		}

		for _, namespace := range output.Namespaces {
			resources = append(resources, &RedshiftServerlessNamespace{
				svc:       svc,
				namespace: namespace,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (n *RedshiftServerlessNamespace) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreationDate", n.namespace.CreationDate).
		Set("NamespaceName", n.namespace.NamespaceName)

	return properties
}

func (n *RedshiftServerlessNamespace) Remove(_ context.Context) error {
	_, err := n.svc.DeleteNamespace(&redshiftserverless.DeleteNamespaceInput{
		NamespaceName: n.namespace.NamespaceName,
	})

	return err
}

func (n *RedshiftServerlessNamespace) String() string {
	return ptr.ToString(n.namespace.NamespaceName)
}
