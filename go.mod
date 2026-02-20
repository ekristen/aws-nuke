module github.com/ekristen/aws-nuke/v3

go 1.24.0

toolchain go1.25.6

require (
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.41.1
	github.com/aws/aws-sdk-go-v2/config v1.32.9
	github.com/aws/aws-sdk-go-v2/credentials v1.19.9
	github.com/aws/aws-sdk-go-v2/service/amp v1.42.5
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.38.4
	github.com/aws/aws-sdk-go-v2/service/appsync v1.53.1
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.19.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.60.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.63.1
	github.com/aws/aws-sdk-go-v2/service/docdb v1.48.9
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.20.9
	github.com/aws/aws-sdk-go-v2/service/dsql v1.12.4
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.290.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.72.0
	github.com/aws/aws-sdk-go-v2/service/efs v1.41.10
	github.com/aws/aws-sdk-go-v2/service/eks v1.80.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.53.2
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.47.0
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.47.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.88.0
	github.com/aws/aws-sdk-go-v2/service/mgn v1.39.1
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.21.17
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.59.3
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.9.19
	github.com/aws/aws-sdk-go-v2/service/s3 v1.96.0
	github.com/aws/aws-sdk-go-v2/service/s3control v1.68.0
	github.com/aws/aws-sdk-go-v2/service/shield v1.34.17
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.8.17
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.6
	github.com/aws/aws-sdk-go-v2/service/textract v1.40.16
	github.com/aws/aws-sdk-go-v2/service/transfer v1.69.1
	github.com/aws/smithy-go v1.24.0
	github.com/ekristen/libnuke v1.3.0
	github.com/fatih/color v1.18.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/gotidy/ptr v1.4.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.6.2
	go.uber.org/ratelimit v0.3.1
	golang.org/x/text v0.33.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.14 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mb0/glob v0.0.0-20160210091149-1eb79d2de6c4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/stevenle/topsort v0.2.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
