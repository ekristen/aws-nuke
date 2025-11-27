module github.com/ekristen/aws-nuke/v3

go 1.24.0

toolchain go1.25.3

require (
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.40.0
	github.com/aws/aws-sdk-go-v2/config v1.28.11
	github.com/aws/aws-sdk-go-v2/credentials v1.17.71
	github.com/aws/aws-sdk-go-v2/service/amp v1.36.0
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.28.12
	github.com/aws/aws-sdk-go-v2/service/appsync v1.42.3
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.44.12
	github.com/aws/aws-sdk-go-v2/service/docdb v1.41.7
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.15.5
	github.com/aws/aws-sdk-go-v2/service/dsql v1.1.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.239.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.54.6
	github.com/aws/aws-sdk-go-v2/service/efs v1.35.4
	github.com/aws/aws-sdk-go-v2/service/eks v1.74.2
	github.com/aws/aws-sdk-go-v2/service/iam v1.38.10
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.44.7
	github.com/aws/aws-sdk-go-v2/service/mgn v1.37.5
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.17.6
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.53.0
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.4.17
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.41.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.72.3
	github.com/aws/aws-sdk-go-v2/service/s3control v1.52.7
	github.com/aws/aws-sdk-go-v2/service/shield v1.34.6
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.3.10
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.1
	github.com/aws/aws-sdk-go-v2/service/transfer v1.55.5
	github.com/aws/smithy-go v1.23.2
	github.com/ekristen/libnuke v1.3.0
	github.com/fatih/color v1.18.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/gotidy/ptr v1.4.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.4.1
	go.uber.org/mock v0.6.0
	go.uber.org/ratelimit v0.3.1
	golang.org/x/text v0.30.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.4 // indirect
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
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
