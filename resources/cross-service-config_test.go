package resources

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// globalServiceClients maps an AWS SDK v2 package name for a global service to the
// filename prefix used by the resources that represent that service. Keep this in sync
// with awsutil.GlobalServices.
var globalServiceClients = map[string]string{
	"iam":        "iam-",
	"route53":    "route53-",
	"cloudfront": "cloudfront",
	"waf":        "waf-",
}

// knownGlobalClientExceptions are files that still build a global service client from
// opts.Config and have not been converted yet. They are broken in exactly the way this
// test describes; the entries exist so the guard can be enforced for everything else in
// the meantime. Do not add to this list.
var knownGlobalClientExceptions = map[string]bool{
	// CloudFormationStack's CreateRoleToDeleteStack setting cannot work in any region
	// for this reason. Converting it is deliberately out of scope for now.
	"cloudformation-stack.go": true,
}

// Test_CrossServiceConfig_GlobalClients guards against building a client for a global
// service straight off opts.Config. The config handed to a lister carries middleware
// that rejects requests to global services so that those resources are only processed
// in the "global" pseudo-region, and that middleware applies to every client built from
// the config, not just the resource's own. A resource that needs a global service to
// support its own removal (creating an IAM role so it can be deleted, say) must build
// that client from opts.CrossServiceConfig() instead, or every call it makes will fail
// with "service 'IAM' is global, but the session is not".
func Test_CrossServiceConfig_GlobalClients(t *testing.T) {
	files, err := filepath.Glob("*.go")
	assert.NoError(t, err)

	patterns := make(map[string]*regexp.Regexp, len(globalServiceClients))
	for pkg := range globalServiceClients {
		patterns[pkg] = regexp.MustCompile(`\b` + pkg + `\.NewFromConfig\(\*opts\.Config\)`)
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") || knownGlobalClientExceptions[file] {
			continue
		}

		contents, err := os.ReadFile(file)
		assert.NoError(t, err)

		for pkg, prefix := range globalServiceClients {
			// The resources that represent the service itself only ever run in the
			// global region, where the config is built to allow it.
			if strings.HasPrefix(file, prefix) {
				continue
			}

			assert.NotRegexp(t, patterns[pkg], string(contents),
				"%s builds a %s client from opts.Config; %s is a global service, so use "+
					"opts.CrossServiceConfig() instead", file, pkg, pkg)
		}
	}
}
