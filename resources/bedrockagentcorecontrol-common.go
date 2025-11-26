package resources

import "slices"

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
