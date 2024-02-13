package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerNotebookInstanceResource = "SageMakerNotebookInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:   SageMakerNotebookInstanceResource,
		Scope:  nuke.Account,
		Lister: &SageMakerNotebookInstanceLister{},
	})
}

type SageMakerNotebookInstanceLister struct{}

func (l *SageMakerNotebookInstanceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListNotebookInstancesInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListNotebookInstances(params)
		if err != nil {
			return nil, err
		}

		for _, notebookInstance := range resp.NotebookInstances {
			resources = append(resources, &SageMakerNotebookInstance{
				svc:                  svc,
				notebookInstanceName: notebookInstance.NotebookInstanceName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerNotebookInstance struct {
	svc                  *sagemaker.SageMaker
	notebookInstanceName *string
}

func (f *SageMakerNotebookInstance) Remove(_ context.Context) error {
	_, err := f.svc.DeleteNotebookInstance(&sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: f.notebookInstanceName,
	})

	return err
}

func (f *SageMakerNotebookInstance) String() string {
	return *f.notebookInstanceName
}
