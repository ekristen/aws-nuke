module github.com/ekristen/aws-nuke/v3

go 1.26.1

require (
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.47.4
	github.com/aws/aws-sdk-go-v2/config v1.33.11
	github.com/aws/aws-sdk-go-v2/credentials v1.20.11
	github.com/aws/aws-sdk-go-v2/service/amp v1.54.7
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.52.3
	github.com/aws/aws-sdk-go-v2/service/appsync v1.62.6
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.74.1
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.75.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.75.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.90.2
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.49.6
	github.com/aws/aws-sdk-go-v2/service/docdb v1.57.5
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.30.0
	github.com/aws/aws-sdk-go-v2/service/dsql v1.22.6
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.341.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.102.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.51.0
	github.com/aws/aws-sdk-go-v2/service/eks v1.105.0
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.52.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.64.5
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.61.5
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.56.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.113.3
	github.com/aws/aws-sdk-go-v2/service/mgn v1.56.6
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.31.5
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.77.1
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.42.6
	github.com/aws/aws-sdk-go-v2/service/ram v1.45.6
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.17.5
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.54.6
	github.com/aws/aws-sdk-go-v2/service/s3 v1.114.5
	github.com/aws/aws-sdk-go-v2/service/s3control v1.79.5
	github.com/aws/aws-sdk-go-v2/service/s3files v1.9.1
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.23.5
	github.com/aws/aws-sdk-go-v2/service/s3vectors v1.16.4
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.85.2
	github.com/aws/aws-sdk-go-v2/service/shield v1.43.6
	github.com/aws/aws-sdk-go-v2/service/sqs v1.52.6
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.17.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.51.5
	github.com/aws/aws-sdk-go-v2/service/textract v1.49.7
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.32.0
	github.com/aws/aws-sdk-go-v2/service/transfer v1.86.2
	github.com/aws/smithy-go v1.28.6
	github.com/ekristen/libnuke v1.3.0
	github.com/fatih/color v1.19.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/gotidy/ptr v1.4.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.10.1
	go.uber.org/mock v0.6.0
	go.uber.org/ratelimit v0.3.1
	golang.org/x/text v0.34.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.24 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.38.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.43.5 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mb0/glob v0.0.0-20160210091149-1eb79d2de6c4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/stevenle/topsort v0.2.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
