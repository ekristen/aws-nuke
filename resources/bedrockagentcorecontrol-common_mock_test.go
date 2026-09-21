package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	agentcoretypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
)

func Test_BedrockAgentCore_HarnessesSupportedInRegion(t *testing.T) {
	a := assert.New(t)

	a.True(harnessesSupportedInRegion("us-east-2"))

	// A region where memories and workload identities live but agent runtimes, and so
	// harnesses, do not.
	a.False(harnessesSupportedInRegion("eu-west-3"))
}

func Test_BedrockAgentCore_ResourceIDFromARN(t *testing.T) {
	a := assert.New(t)

	a.Equal("QA_security_tester-mHkVi5HLjc", resourceIDFromARN(
		ptr.String("arn:aws:bedrock-agentcore:us-east-2:123456789012:memory/QA_security_tester-mHkVi5HLjc")))

	a.Equal("harness_QA_security_tester-bYsDsZHqxV", resourceIDFromARN(
		ptr.String("arn:aws:bedrock-agentcore:us-east-2:123456789012:workload-identity-directory/default/"+
			"workload-identity/harness_QA_security_tester-bYsDsZHqxV")))

	a.Equal("", resourceIDFromARN(nil))
	a.Equal("", resourceIDFromARN(ptr.String("not-an-arn")))
}

func Test_BedrockAgentCore_HarnessManagedResources_AddHarness(t *testing.T) {
	a := assert.New(t)

	managed := &harnessManagedResources{
		AgentRuntimes: map[string]string{},
		Memories:      map[string]string{},
	}

	managed.addHarness(nil)
	a.Empty(managed.AgentRuntimes)
	a.Empty(managed.Memories)

	managed.addHarness(&agentcoretypes.Harness{
		HarnessId: ptr.String("QA_security_tester-VgVWqqFI79"),
		Environment: &agentcoretypes.HarnessEnvironmentProviderMemberAgentCoreRuntimeEnvironment{
			Value: agentcoretypes.HarnessAgentCoreRuntimeEnvironment{
				AgentRuntimeId: ptr.String("harness_QA_security_tester-bYsDsZHqxV"),
			},
		},
		Memory: &agentcoretypes.HarnessMemoryConfigurationMemberManagedMemoryConfiguration{
			Value: agentcoretypes.HarnessManagedMemoryConfiguration{
				Arn: ptr.String("arn:aws:bedrock-agentcore:us-east-2:123456789012:memory/" +
					"QA_security_tester-mHkVi5HLjc"),
			},
		},
	})

	a.Equal(map[string]string{
		"harness_QA_security_tester-bYsDsZHqxV": "QA_security_tester-VgVWqqFI79",
	}, managed.AgentRuntimes)
	a.Equal(map[string]string{
		"QA_security_tester-mHkVi5HLjc": "QA_security_tester-VgVWqqFI79",
	}, managed.Memories)
}

// A memory the caller created and pointed the harness at is not owned by the harness,
// so it stays deletable on its own.
func Test_BedrockAgentCore_HarnessManagedResources_UnmanagedMemory(t *testing.T) {
	a := assert.New(t)

	managed := &harnessManagedResources{
		AgentRuntimes: map[string]string{},
		Memories:      map[string]string{},
	}

	managed.addHarness(&agentcoretypes.Harness{
		HarnessId: ptr.String("QA_security_tester-VgVWqqFI79"),
		Memory: &agentcoretypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration{
			Value: agentcoretypes.HarnessAgentCoreMemoryConfiguration{
				Arn: ptr.String("arn:aws:bedrock-agentcore:us-east-2:123456789012:memory/byo-mem-abc123"),
			},
		},
	})

	a.Empty(managed.Memories)
}

func Test_BedrockAgentCore_RegistryWorkloadIdentityName(t *testing.T) {
	a := assert.New(t)

	a.Equal("registry-Olb0p1W8eCXAG35A", registryWorkloadIdentityName("Olb0p1W8eCXAG35A"))
}
