package awsutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	credentialsv2 "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go/aws"                      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/credentials"          //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/endpoints"            //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/request"              //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/session"              //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iottwinmaker"     //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/s3control"        //nolint:staticcheck

	liberrors "github.com/ekristen/libnuke/pkg/errors"

	"github.com/ekristen/aws-nuke/v3/pkg/config"
)

const (
	GlobalRegionID = "global"
)

var (
	// DefaultRegionID The default region. Can be customized for non AWS implementations
	DefaultRegionID = "us-east-1"

	// DefaultAWSPartitionID The default aws partition. Can be customized for non AWS implementations
	DefaultAWSPartitionID = "aws"
)

type Credentials struct {
	Profile string

	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string //nolint:gosec // not a hardcoded credential
	AssumeRoleArn   string
	ExternalID      string
	RoleSessionName string

	Credentials *credentials.Credentials

	CustomEndpoints config.CustomEndpoints
	session         *session.Session
	cfg             *awsv2.Config
}

func (c *Credentials) HasProfile() bool {
	return strings.TrimSpace(c.Profile) != ""
}

func (c *Credentials) HasAwsCredentials() bool {
	return c.Credentials != nil
}

func (c *Credentials) HasKeys() bool {
	return strings.TrimSpace(c.AccessKeyID) != "" ||
		strings.TrimSpace(c.SecretAccessKey) != "" ||
		strings.TrimSpace(c.SessionToken) != ""
}

func (c *Credentials) Validate() error {
	if c.HasProfile() && c.HasKeys() {
		return fmt.Errorf("specify either the --profile flag or " +
			"--access-key-id with --secret-access-key and optionally " +
			"--session-token, but not both")
	}

	return nil
}

// FUTURE(187): when all services are migrated to SDK v2, remove usage of
// session.Session throughout
func (c *Credentials) rootSession() (*session.Session, error) {
	if c.session == nil {
		var opts session.Options

		region := DefaultRegionID
		log.Debugf("creating new root session in %s", region)

		switch {
		case c.HasAwsCredentials():
			opts = session.Options{
				Config: aws.Config{
					Credentials: c.Credentials,
				},
			}
		case c.HasProfile() && c.HasKeys():
			return nil, fmt.Errorf("you have to specify a profile or credentials for at least one region")

		case c.HasKeys():
			opts = session.Options{
				Config: aws.Config{
					Credentials: c.awsNewStaticCredentials(),
				},
			}

		case c.HasProfile():
			fallthrough //nolint:gocritic

		default:
			opts = session.Options{
				SharedConfigState:       session.SharedConfigEnable,
				Profile:                 c.Profile,
				AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
			}
		}

		opts.Config.Region = aws.String(region)
		opts.Config.DisableRestProtocolURICleaning = aws.Bool(true)

		sess, err := session.NewSessionWithOptions(opts)
		if err != nil {
			return nil, err
		}

		// if given a role to assume, overwrite the session credentials with assume role credentials
		if c.AssumeRoleArn != "" {
			sess.Config.Credentials = stscreds.NewCredentials(sess, c.AssumeRoleArn, func(p *stscreds.AssumeRoleProvider) {
				if c.RoleSessionName != "" {
					p.RoleSessionName = c.RoleSessionName
				}

				if c.ExternalID != "" {
					p.ExternalID = aws.String(c.ExternalID)
				}
			})
		}

		c.session = sess
	}

	return c.session, nil
}

func (c *Credentials) awsNewStaticCredentials() *credentials.Credentials {
	if !c.HasKeys() {
		return credentials.NewEnvCredentials()
	}
	return credentials.NewStaticCredentials(
		strings.TrimSpace(c.AccessKeyID),
		strings.TrimSpace(c.SecretAccessKey),
		strings.TrimSpace(c.SessionToken),
	)
}

func (c *Credentials) awsNewStaticCredentialsV2() awsv2.CredentialsProvider {
	if !c.HasKeys() {
		return envCredentialsProviderV2{}
	}
	return credentialsv2.NewStaticCredentialsProvider(
		strings.TrimSpace(c.AccessKeyID),
		strings.TrimSpace(c.SecretAccessKey),
		strings.TrimSpace(c.SessionToken),
	)
}

func (c *Credentials) NewSession(region, serviceType string) (*session.Session, error) {
	log.Debugf("creating new session in %s for %s", region, serviceType)

	global := false

	if region == GlobalRegionID {
		region = DefaultRegionID
		global = true
	}

	var sess *session.Session
	isCustom := false
	if customRegion := c.CustomEndpoints.GetRegion(region); customRegion != nil {
		customService := customRegion.Services.GetService(serviceType)
		if customService == nil {
			return nil, liberrors.ErrSkipRequest(fmt.Sprintf(
				".service '%s' is not available in region '%s'",
				serviceType, region))
		}
		conf := &aws.Config{
			Region:      &region,
			Endpoint:    &customService.URL,
			Credentials: c.awsNewStaticCredentials(),
		}
		if customService.TLSInsecureSkipVerify {
			conf.HTTPClient = &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			}}
		}

		var err error
		sess, err = session.NewSession(conf)
		if err != nil {
			return nil, err
		}

		isCustom = true
	}

	if sess == nil {
		root, err := c.rootSession()
		if err != nil {
			return nil, err
		}

		sess = root.Copy(&aws.Config{
			Region: &region,
		})
	}

	sess.Handlers.Send.PushFront(func(r *request.Request) {
		log.Tracef("sending AWS request:\n%s", DumpRequest(r.HTTPRequest))
	})

	sess.Handlers.ValidateResponse.PushFront(func(r *request.Request) {
		log.Tracef("received AWS response:\n%s", DumpResponse(r.HTTPResponse))
	})

	if !isCustom {
		sess.Handlers.Validate.PushFront(skipMissingServiceInRegionHandler)
		sess.Handlers.Validate.PushFront(skipGlobalHandler(global))
	}
	return sess, nil
}

func skipMissingServiceInRegionHandler(r *request.Request) {
	region := *r.Config.Region
	service := r.ClientInfo.ServiceName

	rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), DefaultAWSPartitionID, service)
	if !ok {
		// This means that the service does not exist and this shouldn't be handled here.
		return
	}

	if len(rs) == 0 {
		// Avoid to throw an error on global services.
		return
	}

	_, ok = rs[region]
	if !ok {
		r.Error = liberrors.ErrSkipRequest(fmt.Sprintf(
			"service '%s' is not available in region '%s'",
			service, region))
	}
}

func skipGlobalHandler(global bool) func(r *request.Request) {
	return func(r *request.Request) {
		service := r.ClientInfo.ServiceName
		if service == s3control.ServiceName {
			service = s3control.EndpointsID
			// Rewrite S3 Control ServiceName to proper EndpointsID
			// https://github.com/rebuy-de/aws-nuke/issues/708
		}
		if service == iottwinmaker.ServiceName {
			service = iottwinmaker.EndpointsID
			// IoTTwinMaker have two endpoints, must point on "api" one
			// https://docs.aws.amazon.com/iot-twinmaker/latest/guide/endpionts-and-quotas.html
		}
		rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), DefaultAWSPartitionID, service)
		if !ok {
			// This means that the service does not exist in the endpoints list.
			if global {
				r.Error = liberrors.ErrSkipRequest(
					fmt.Sprintf("service '%s' is was not found in the endpoint list; assuming it is not global",
						service))
			} else {
				host := r.HTTPRequest.URL.Hostname()
				_, err := net.DefaultResolver.LookupHost(r.Context(), host)
				if err != nil {
					log.Debug(err)
					r.Error = liberrors.ErrUnknownEndpoint(
						fmt.Sprintf("DNS lookup failed for %s; assuming it does not exist in this region", host))
				}
			}
			return
		}

		if len(rs) == 0 && !global {
			r.Error = liberrors.ErrSkipRequest(
				fmt.Sprintf("service '%s' is global, but the session is not", service))
			return
		}

		if (len(rs) > 0 && global) && service != "sts" {
			r.Error = liberrors.ErrSkipRequest(
				fmt.Sprintf("service '%s' is not global, but the session is", service))
			return
		}
	}
}

// SDK v2 does not directly expose an environment creds provider since that
// functionality has been opaquely rolled into LoadDefaultConfig
//
// this provider recreates the behavior that v1 had (including support for
// extra nonstandard envs)
type envCredentialsProviderV2 struct{}

func (envCredentialsProviderV2) Retrieve(ctx context.Context) (awsv2.Credentials, error) {
	id := os.Getenv("AWS_ACCESS_KEY_ID")
	if id == "" {
		id = os.Getenv("AWS_ACCESS_KEY")
	}

	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secret == "" {
		secret = os.Getenv("AWS_SECRET_KEY")
	}

	if id == "" {
		return awsv2.Credentials{}, fmt.Errorf("AWS access key unset")
	}

	if secret == "" {
		return awsv2.Credentials{}, fmt.Errorf("AWS secret key unset")
	}

	return awsv2.Credentials{
		AccessKeyID:     id,
		SecretAccessKey: secret,
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
	}, nil
}
