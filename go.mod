module github.com/ekristen/aws-nuke/v3

go 1.26.1

require (
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.48.0
	github.com/aws/aws-sdk-go-v2/config v1.33.18
	github.com/aws/aws-sdk-go-v2/credentials v1.21.0
	github.com/aws/aws-sdk-go-v2/service/amp v1.55.4
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.52.9
	github.com/aws/aws-sdk-go-v2/service/appsync v1.62.12
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.74.7
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.80.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.75.6
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.92.2
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.50.0
	github.com/aws/aws-sdk-go-v2/service/docdb v1.59.1
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.30.6
	github.com/aws/aws-sdk-go-v2/service/dsql v1.22.12
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.347.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.105.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.51.6
	github.com/aws/aws-sdk-go-v2/service/eks v1.105.6
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.52.6
	github.com/aws/aws-sdk-go-v2/service/fis v1.46.12
	github.com/aws/aws-sdk-go-v2/service/iam v1.66.0
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.62.5
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.58.0
	github.com/aws/aws-sdk-go-v2/service/lambda v1.115.0
	github.com/aws/aws-sdk-go-v2/service/mgn v1.57.3
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.31.11
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.80.0
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.43.5
	github.com/aws/aws-sdk-go-v2/service/ram v1.46.1
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.19.0
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.54.12
	github.com/aws/aws-sdk-go-v2/service/s3 v1.114.10
	github.com/aws/aws-sdk-go-v2/service/s3control v1.80.1
	github.com/aws/aws-sdk-go-v2/service/s3files v1.10.5
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.26.0
	github.com/aws/aws-sdk-go-v2/service/s3vectors v1.17.5
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.87.0
	github.com/aws/aws-sdk-go-v2/service/shield v1.44.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.53.0
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.17.12
	github.com/aws/aws-sdk-go-v2/service/sts v1.52.0
	github.com/aws/aws-sdk-go-v2/service/textract v1.50.0
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.33.4
	github.com/aws/aws-sdk-go-v2/service/transfer v1.86.8
	github.com/aws/smithy-go v1.28.9
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
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.27 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.39.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.44.5 // indirect
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
