package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/eks" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EKSFargateProfileResource = "EKSFargateProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:     EKSFargateProfileResource,
		Scope:    nuke.Account,
		Resource: &EKSFargateProfile{},
		Lister:   &EKSFargateProfileLister{},
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
				profResp, err := svc.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
					ClusterName:        clusterName,
					FargateProfileName: name,
				})
				if err != nil {
					logrus.WithError(err).Error("unable to describe fargate profile")
					continue
				}

				resources = append(resources, &EKSFargateProfile{
					svc:       svc,
					Name:      name,
					Cluster:   clusterName,
					CreatedAt: profResp.FargateProfile.CreatedAt,
					Tags:      profResp.FargateProfile.Tags,
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
	svc       *eks.EKS
	Cluster   *string
	Name      *string
	CreatedAt *time.Time
	Tags      map[string]*string
}

func (r *EKSFargateProfile) Remove(_ context.Context) error {
	_, err := r.svc.DeleteFargateProfile(&eks.DeleteFargateProfileInput{
		ClusterName:        r.Cluster,
		FargateProfileName: r.Name,
	})
	return err
}

func (r *EKSFargateProfile) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EKSFargateProfile) String() string {
	return fmt.Sprintf("%s:%s", *r.Cluster, *r.Name)
}
