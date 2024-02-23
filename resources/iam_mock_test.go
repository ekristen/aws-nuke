//go:generate ../mocks/generate_mocks.sh iam iamiface
package resources

import _ "github.com/golang/mock/mockgen"

// Note: empty on purpose, this file exist purely to generate mocks for the IAM service
