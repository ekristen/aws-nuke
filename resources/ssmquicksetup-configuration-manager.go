package resources

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/ssmquicksetup"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SSMQuickSetupConfigurationManagerResource = "SSMQuickSetupConfigurationManager"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSMQuickSetupConfigurationManagerResource,
		Scope:    nuke.Account,
		Resource: &SSMQuickSetupConfigurationManager{},
		Lister:   &SSMQuickSetupConfigurationManagerLister{},
		Settings: []string{
			"CreateRole",
		},
	})
}

type SSMQuickSetupConfigurationManagerLister struct{}

func (l *SSMQuickSetupConfigurationManagerLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ssmquicksetup.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	res, err := svc.ListConfigurationManagers(ctx, &ssmquicksetup.ListConfigurationManagersInput{})
	if err != nil {
		return nil, err
	}

	for _, p := range res.ConfigurationManagersList {
		resources = append(resources, &SSMQuickSetupConfigurationManager{
			svc:    svc,
			iamSvc: iam.NewFromConfig(*opts.Config),
			stsSvc: sts.NewFromConfig(*opts.Config),
			ARN:    p.ManagerArn,
			Name:   p.Name,
		})
	}

	return resources, nil
}

type SSMQuickSetupConfigurationManager struct {
	svc      *ssmquicksetup.Client
	iamSvc   *iam.Client
	stsSvc   *sts.Client
	ARN      *string
	Name     *string
	settings *settings.Setting
}

// GetName returns the name of the resource or the last part of the ARN if not set so that the stringer resource has
// a value to display
func (r *SSMQuickSetupConfigurationManager) GetName() string {
	if ptr.ToString(r.Name) != "" {
		return ptr.ToString(r.Name)
	}

	parts := strings.Split(ptr.ToString(r.ARN), "/")
	return parts[len(parts)-1]
}

func (r *SSMQuickSetupConfigurationManager) Settings(setting *settings.Setting) {
	r.settings = setting
}

func (r *SSMQuickSetupConfigurationManager) Remove(ctx context.Context) error {
	// Try to delete first, then create role if we get the specific error
	_, err := r.svc.DeleteConfigurationManager(ctx, &ssmquicksetup.DeleteConfigurationManagerInput{
		ManagerArn: r.ARN,
	})

	// Check if we got the specific error about role access and if CreateRole setting is enabled
	if err != nil && r.settings.GetBool("CreateRole") {
		if roleName := r.extractRoleNameFromError(err); roleName != "" {
			// Create the specific role mentioned in the error message
			if createErr := r.createRoleFromError(ctx, roleName); createErr != nil {
				return createErr
			}
			// Retry the deletion after creating role
			_, err = r.svc.DeleteConfigurationManager(ctx, &ssmquicksetup.DeleteConfigurationManagerInput{
				ManagerArn: r.ARN,
			})
		}
	}

	return err
}

func (r *SSMQuickSetupConfigurationManager) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SSMQuickSetupConfigurationManager) String() string {
	return r.GetName()
}

// extractRoleNameFromError extracts the role name from the error message
func (r *SSMQuickSetupConfigurationManager) extractRoleNameFromError(err error) string {
	errStr := err.Error()

	// Look for the pattern "Role ROLE_NAME can't be accessed"
	if strings.Contains(errStr, "can't be accessed") {
		// Find "Role " and extract the role name after it
		roleIndex := strings.Index(errStr, "Role ")
		if roleIndex != -1 {
			roleStart := roleIndex + 5 // Length of "Role "
			roleEnd := strings.Index(errStr[roleStart:], " can't be accessed")
			if roleEnd != -1 {
				return errStr[roleStart : roleStart+roleEnd]
			}
		}
	}

	return ""
}

// createRoleFromError creates the specific role mentioned in the error message
func (r *SSMQuickSetupConfigurationManager) createRoleFromError(ctx context.Context, roleName string) error {
	// Get current account ID
	callerIdentity, err := r.stsSvc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	accountID := *callerIdentity.Account

	// Determine which type of role to create based on the role name
	if strings.Contains(roleName, "Administration") {
		return r.createAdminRole(ctx, roleName, accountID)
	} else if strings.Contains(roleName, "Execution") {
		return r.createExecRole(ctx, roleName, accountID)
	}

	return nil
}

// createAdminRole creates the LocalAdministrationRole with CloudFormation trust policy
func (r *SSMQuickSetupConfigurationManager) createAdminRole(ctx context.Context, roleName, accountID string) error {
	// Define trust policy for CloudFormation service with conditions
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Service": "cloudformation.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
				"Condition": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"aws:SourceAccount": accountID,
					},
					"StringLike": map[string]interface{}{
						"aws:SourceArn": "arn:aws:cloudformation:*:" + accountID + ":stackset/AWS-QuickSetup-*",
					},
				},
			},
		},
	}

	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return err
	}

	// Create the role
	_, err = r.iamSvc.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
		Description:              aws.String("LocalAdministrationRole created by aws-nuke for SSM QuickSetup Configuration Manager deletion"),
		Path:                     aws.String("/"),
		Tags: []iamtypes.Tag{
			{
				Key:   aws.String("CreatedBy"),
				Value: aws.String("aws-nuke"),
			},
			{
				Key:   aws.String("Purpose"),
				Value: aws.String("SSMQuickSetupConfigurationManager-Deletion"),
			},
			{
				Key:   aws.String("ConfigurationManager"),
				Value: aws.String(r.GetName()),
			},
		},
	})

	if err != nil {
		return err
	}

	// Create inline policy for assuming execution roles (both LocalExecutionRole and LocalDeploymentExecutionRole)
	execRoleBaseName := strings.Replace(roleName, "LocalAdministrationRole", "", 1)
	policyDocument := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Action": []string{"sts:AssumeRole"},
				"Resource": []string{
					"arn:aws:iam::" + accountID + ":role/" + execRoleBaseName + "LocalExecutionRole",
					"arn:aws:iam::" + accountID + ":role/" + execRoleBaseName + "LocalDeploymentExecutionRole",
				},
				"Effect": "Allow",
			},
		},
	}

	policyJSON, err := json.Marshal(policyDocument)
	if err != nil {
		return err
	}

	// Attach inline policy
	_, err = r.iamSvc.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String("AssumeLocalExecutionRole"),
		PolicyDocument: aws.String(string(policyJSON)),
	})

	return err
}

// createExecRole creates the LocalExecutionRole/LocalDeploymentExecutionRole with LocalAdministrationRole trust policy
func (r *SSMQuickSetupConfigurationManager) createExecRole(ctx context.Context, roleName, accountID string) error {
	// Derive the corresponding admin role name from the execution role name
	var adminRoleName string
	if strings.Contains(roleName, "LocalExecutionRole") {
		adminRoleName = strings.Replace(roleName, "LocalExecutionRole", "LocalAdministrationRole", 1)
	} else if strings.Contains(roleName, "LocalDeploymentExecutionRole") {
		adminRoleName = strings.Replace(roleName, "LocalDeploymentExecutionRole", "LocalAdministrationRole", 1)
	}

	// Define trust policy for LocalAdministrationRole
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"AWS": "arn:aws:iam::" + accountID + ":role/" + adminRoleName,
				},
				"Action": "sts:AssumeRole",
			},
		},
	}

	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return err
	}

	// Create the role
	_, err = r.iamSvc.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
		Description:              aws.String("LocalExecutionRole created by aws-nuke for SSM QuickSetup Configuration Manager deletion"),
		Path:                     aws.String("/"),
		Tags: []iamtypes.Tag{
			{
				Key:   aws.String("CreatedBy"),
				Value: aws.String("aws-nuke"),
			},
			{
				Key:   aws.String("Purpose"),
				Value: aws.String("SSMQuickSetupConfigurationManager-Deletion"),
			},
			{
				Key:   aws.String("ConfigurationManager"),
				Value: aws.String(r.GetName()),
			},
		},
	})

	if err != nil {
		return err
	}

	// Attach the AWSQuickSetupDeploymentRolePolicy
	policyArn := "arn:aws:iam::aws:policy/AWSQuickSetupDeploymentRolePolicy"
	_, err = r.iamSvc.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})

	return err
}
