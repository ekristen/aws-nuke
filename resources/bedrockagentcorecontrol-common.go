package resources

import (
	"context"
	"slices"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	agentcoretypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
)

// Note: if any regions are commented out that means they are not actually supported
// contrary to what the documentation says.
var (
	SupportedRegions = []string{
		"us-east-1",      // US East (N. Virginia)
		"us-east-2",      // US East (Ohio)
		"us-west-2",      // US West (Oregon)
		"ap-southeast-2", // Asia Pacific (Sydney)
		"ap-south-1",     // Asia Pacific (Mumbai)
		"ap-northeast-1", // Asia Pacific (Tokyo)
		"ap-southeast-1", // Asia Pacific (Singapore)
		"ap-northeast-2", // Asia Pacific (Seoul)
		"eu-west-1",      // Europe (Ireland)
		"eu-central-1",   // Europe (Frankfurt)
		"eu-north-1",     // Europe (Stockholm)
		"eu-west-2",      // Europe (London)
		"eu-west-3",      // Europe (Paris)
		"ca-central-1",   // Canada (Central)
	}

	AgentRuntimeSupportedRegions = []string{
		"us-east-1",      // US East (N. Virginia)
		"us-east-2",      // US East (Ohio)
		"us-west-2",      // US West (Oregon)
		"ap-southeast-2", // Asia Pacific (Sydney)
		"ap-south-1",     // Asia Pacific (Mumbai)
		"ap-northeast-1", // Asia Pacific (Tokyo)
		"ap-southeast-1", // Asia Pacific (Singapore)
		// "ap-northeast-2", // Asia Pacific (Seoul)
		"eu-west-1",    // Europe (Ireland)
		"eu-central-1", // Europe (Frankfurt)
		// "eu-north-1",     // Europe (Stockholm)
		// "eu-west-2",      // Europe (London)
		// "eu-west-3",      // Europe (Paris)
		// "ca-central-1",   // Canada (Central)
	}

	BuiltInToolsSupportedRegions = []string{
		"us-east-1",      // US East (N. Virginia)
		"us-east-2",      // US East (Ohio)
		"us-west-2",      // US West (Oregon)
		"ap-southeast-2", // Asia Pacific (Sydney)
		"ap-south-1",     // Asia Pacific (Mumbai)
		"ap-northeast-1", // Asia Pacific (Tokyo)
		"ap-southeast-1", // Asia Pacific (Singapore)
		// "ap-northeast-2", // Asia Pacific (Seoul)
		"eu-west-1",    // Europe (Ireland)
		"eu-central-1", // Europe (Frankfurt)
		// "eu-north-1",     // Europe (Stockholm)
		// "eu-west-2",      // Europe (London)
		// "eu-west-3",      // Europe (Paris)
		// "ca-central-1",   // Canada (Central)
	}
)

// BedrockAgentCoreControlLister is a common struct that can be embedded in all
// Bedrock AgentCore Control resource listers to provide region support checking.
type BedrockAgentCoreControlLister struct {
	supportedRegions []string
}

func (l *BedrockAgentCoreControlLister) SetSupportedRegions(regions []string) {
	l.supportedRegions = regions
}

func (l *BedrockAgentCoreControlLister) IsSupportedRegion(region string) bool {
	// ref: https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/agentcore-regions.html
	if len(l.supportedRegions) == 0 {
		l.supportedRegions = SupportedRegions
	}

	return slices.Contains(l.supportedRegions, region)
}

// harnessesSupportedInRegion reports whether a harness can exist in the region. A
// harness runs on an agent runtime, so it is confined to the regions those are.
func harnessesSupportedInRegion(region string) bool {
	return slices.Contains(AgentRuntimeSupportedRegions, region)
}

// harnessManagedResources holds the resources that harnesses create and own on the
// account's behalf. AWS refuses to delete any of them directly -- DeleteAgentRuntime,
// DeleteMemory and DeleteWorkloadIdentity all come back with a ValidationException
// saying to delete the managing resource instead -- so the listers filter them out and
// leave them for the harness to take with it.
type harnessManagedResources struct {
	// AgentRuntimes maps an agent runtime id to the id of the harness that owns it.
	AgentRuntimes map[string]string
	// Memories maps a memory id to the id of the harness that owns it.
	Memories map[string]string
}

// listHarnessManagedResources walks the harnesses in the region and records what each
// one owns. It is best effort: if the harnesses cannot be listed there is nothing to
// filter against and the resources are handled as if they were standalone, which is how
// they behaved before harnesses existed.
func listHarnessManagedResources(ctx context.Context, svc *bedrockagentcorecontrol.Client,
	logger *logrus.Entry) *harnessManagedResources {
	managed := &harnessManagedResources{
		AgentRuntimes: map[string]string{},
		Memories:      map[string]string{},
	}

	params := &bedrockagentcorecontrol.ListHarnessesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListHarnessesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			logger.Warnf("unable to list harnesses to determine managed resources: %v", err)
			return managed
		}

		for _, summary := range resp.Harnesses {
			getResp, err := svc.GetHarness(ctx, &bedrockagentcorecontrol.GetHarnessInput{
				HarnessId: summary.HarnessId,
			})
			if err != nil {
				logger.Warnf("unable to fetch harness %s: %v", ptr.ToString(summary.HarnessId), err)
				continue
			}

			managed.addHarness(getResp.Harness)
		}
	}

	return managed
}

// addHarness records the resources a single harness owns.
func (m *harnessManagedResources) addHarness(harness *agentcoretypes.Harness) {
	if harness == nil {
		return
	}

	harnessID := ptr.ToString(harness.HarnessId)

	if env, ok := harness.Environment.(*agentcoretypes.HarnessEnvironmentProviderMemberAgentCoreRuntimeEnvironment); ok {
		if runtimeID := ptr.ToString(env.Value.AgentRuntimeId); runtimeID != "" {
			m.AgentRuntimes[runtimeID] = harnessID
		}
	}

	// Only a managed memory belongs to the harness. A memory the caller created and
	// pointed the harness at stays theirs to delete.
	if memory, ok := harness.Memory.(*agentcoretypes.HarnessMemoryConfigurationMemberManagedMemoryConfiguration); ok {
		if memoryID := resourceIDFromARN(memory.Value.Arn); memoryID != "" {
			m.Memories[memoryID] = harnessID
		}
	}
}

// resourceIDFromARN returns the last path segment of an ARN's resource part, which for
// the AgentCore resources is the id or name that the list and delete calls use.
func resourceIDFromARN(value *string) string {
	parsed, err := arn.Parse(ptr.ToString(value))
	if err != nil {
		return ""
	}

	if idx := strings.LastIndex(parsed.Resource, "/"); idx != -1 {
		return parsed.Resource[idx+1:]
	}

	return parsed.Resource
}
