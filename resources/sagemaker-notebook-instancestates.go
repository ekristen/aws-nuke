package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerNotebookInstanceStateResource = "SageMakerNotebookInstanceState"

func init() {
	registry.Register(&registry.Registration{
		Name:   SageMakerNotebookInstanceStateResource,
		Scope:  nuke.Account,
		Lister: &SageMakerNotebookInstanceStateLister{},
	})
}

type SageMakerNotebookInstanceStateLister struct{}

func (l *SageMakerNotebookInstanceStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &SageMakerNotebookInstanceState{
				svc:                  svc,
				notebookInstanceName: notebookInstance.NotebookInstanceName,
				instanceStatus:       notebookInstance.NotebookInstanceStatus,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerNotebookInstanceState struct {
	svc                  *sagemaker.SageMaker
	notebookInstanceName *string
	instanceStatus       *string
}

func (f *SageMakerNotebookInstanceState) Remove(_ context.Context) error {

	_, err := f.svc.StopNotebookInstance(&sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: f.notebookInstanceName,
	})

	return err
}

func (f *SageMakerNotebookInstanceState) String() string {
	return *f.notebookInstanceName
}

func (f *SageMakerNotebookInstanceState) Filter() error {
	if strings.ToLower(*f.instanceStatus) == "stopped" {
		return fmt.Errorf("already stopped")
	}
	return nil
}
