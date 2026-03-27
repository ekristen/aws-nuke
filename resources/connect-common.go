package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/connect"
	connecttypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
)

func listConnectInstances(ctx context.Context, svc *connect.Client) ([]connecttypes.InstanceSummary, error) {
	var instances []connecttypes.InstanceSummary

	params := &connect.ListInstancesInput{
		MaxResults: nil, // API max is 10, which is the default
	}

	paginator := connect.NewListInstancesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		instances = append(instances, resp.InstanceSummaryList...)
	}

	return instances, nil
}
