package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EKSFargateProfileResource = "EKSFargateProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:   EKSFargateProfileResource,
		Scope:  nuke.Account,
		Lister: &EKSFargateProfileLister{},
		DeprecatedAliases: []string{
			"EKSFargateProfiles",
		},
	})
}

type EKSFargateProfileLister struct{}

func (l *EKSFargateProfileLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := eks.New(opts.Session)
	var clusterNames []*string
	var resources []resource.Resource

	clusterInputParams := &eks.ListClustersInput{
		MaxResults: aws.Int64(100),
	}

	// fetch all cluster names
	for {
		resp, err := svc.ListClusters(clusterInputParams)
		if err != nil {
			return nil, err
		}

		clusterNames = append(clusterNames, resp.Clusters...)

		if resp.NextToken == nil {
			break
		}

		clusterInputParams.NextToken = resp.NextToken
	}

	fargateInputParams := &eks.ListFargateProfilesInput{
		MaxResults: aws.Int64(100),
	}

	// fetch the associated eks fargate profiles
	for _, clusterName := range clusterNames {
		fargateInputParams.ClusterName = clusterName

		for {
			resp, err := svc.ListFargateProfiles(fargateInputParams)
			if err != nil {
				return nil, err
			}

			for _, name := range resp.FargateProfileNames {
				resources = append(resources, &EKSFargateProfile{
					svc:     svc,
					name:    name,
					cluster: clusterName,
				})
			}

			if resp.NextToken == nil {
				fargateInputParams.NextToken = nil
				break
			}

			fargateInputParams.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

type EKSFargateProfile struct {
	svc     *eks.EKS
	cluster *string
	name    *string
}

func (fp *EKSFargateProfile) Remove(_ context.Context) error {
	_, err := fp.svc.DeleteFargateProfile(&eks.DeleteFargateProfileInput{
		ClusterName:        fp.cluster,
		FargateProfileName: fp.name,
	})
	return err
}

func (fp *EKSFargateProfile) Properties() types.Properties {
	return types.NewProperties().
		Set("Cluster", *fp.cluster).
		Set("Profile", *fp.name)
}

func (fp *EKSFargateProfile) String() string {
	return fmt.Sprintf("%s:%s", *fp.cluster, *fp.name)
}
