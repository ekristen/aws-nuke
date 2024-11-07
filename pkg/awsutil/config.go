package awsutil

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	liberrors "github.com/ekristen/libnuke/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *Credentials) NewConfig(ctx context.Context, region, serviceType string) (*aws.Config, error) {
	log.Debugf("creating new config in %s for %s", region, serviceType)

	var global bool
	if region == GlobalRegionID {
		region = "us-east-1"
		global = true
	}

	var cfg *aws.Config
	if customRegion := c.CustomEndpoints.GetRegion(region); customRegion != nil {
		var opts []func(*config.LoadOptions) error

		customService := customRegion.Services.GetService(serviceType)
		if customService == nil {
			return nil, liberrors.ErrSkipRequest(fmt.Sprintf(
				".service '%s' is not available in region '%s'",
				serviceType, region))
		}

		opts = append(opts,
			config.WithRegion(region),
			config.WithCredentialsProvider(c.awsNewStaticCredentialsV2()),
			config.WithBaseEndpoint(customService.URL))

		if customService.TLSInsecureSkipVerify {
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				},
			}
			opts = append(opts, config.WithHTTPClient(client))
		}

		cfgv, err := config.LoadDefaultConfig(ctx, opts...)
		if err != nil {
			return nil, err
		}

		cfg = &cfgv
	}

	if cfg == nil {
		root, err := c.rootConfig(ctx)
		if err != nil {
			return nil, err
		}

		cfgCopy := root.Copy()
		cfgCopy.Region = region
		if global {
			cfgCopy.APIOptions = append(cfgCopy.APIOptions, func(stack *middleware.Stack) error {
				return stack.Initialize.Add(SkipGlobal{}, middleware.After)
			})
		}
		cfg = &cfgCopy
	}

	return cfg, nil
}

func (c *Credentials) rootConfig(ctx context.Context) (*aws.Config, error) {
	if c.cfg != nil {
		return c.cfg, nil
	}

	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithAPIOptions([]func(*middleware.Stack) error{
		func(stack *middleware.Stack) error {
			return errors.Join(
				stack.Finalize.Add(traceRequest{}, middleware.After),
				stack.Deserialize.Add(traceResponse{}, middleware.After),
			)
		},
	}))

	region := DefaultRegionID
	log.Debugf("creating new root session in %s", region)

	switch {
	case c.HasAwsCredentials(): // adapts from v1 credentials provider
		creds, err := c.Credentials.GetWithContext(ctx)
		if err != nil {
			return nil, err
		}

		static := credentials.NewStaticCredentialsProvider(
			creds.AccessKeyID,
			creds.SecretAccessKey,
			creds.SessionToken,
		)
		opts = append(opts, config.WithCredentialsProvider(static))

	case c.HasProfile() && c.HasKeys():
		return nil, fmt.Errorf("you have to specify a profile or credentials for at least one region")

	case c.HasKeys():
		static := credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(c.AccessKeyID),
			strings.TrimSpace(c.SecretAccessKey),
			strings.TrimSpace(c.SessionToken),
		)
		opts = append(opts, config.WithCredentialsProvider(static))

	case c.HasProfile():
		fallthrough //nolint:gocritic

	default:
		opts = append(opts, config.WithSharedConfigProfile(c.Profile))
	}

	opts = append(opts, config.WithRegion(region))
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// if given a role to assume, overwrite the session credentials with assume role credentials
	if c.AssumeRoleArn != "" {
		cfg.Credentials = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), c.AssumeRoleArn, func(p *stscreds.AssumeRoleOptions) {
			if c.RoleSessionName != "" {
				p.RoleSessionName = c.RoleSessionName
			}

			if c.ExternalID != "" {
				p.ExternalID = aws.String(c.ExternalID)
			}
		})
	}

	c.cfg = &cfg
	return c.cfg, nil
}

// SkipGlobal returns ErrSkipRequest when operating in the global
// pseudo-region.
//
// FUTURE: define mechanism for allowing specific resources, such as those in
// IAM, to override this skip.
//
// The simplest way to do this would be to remove this middleware through
// functional options on relevant operations. e.g.:
//
//	 func isGlobalResource(o *iam.Options) {
//		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
//			stack.Initialize.Remove(config.SkipGlobal{}.ID())
//		})
//	 }
//
//	 // per-operation
//	 out, err := svc.ListRoles(context.Background(), nil, isGlobalResource)
//	 // on a client, if you know you're only operating in the context of global resources
//	 svc := iam.NewFromConfig(cfg, isGlobalResource)
//
// You could also define some kind of "is global resource" Context flag, which
// SkipGlobal could react to. That may be preferrable to having SkipGlobal be
// exported from this package.
type SkipGlobal struct{}

func (SkipGlobal) ID() string {
	return "aws-nuke::skipGlobal"
}

func (SkipGlobal) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, md middleware.Metadata, err error,
) {
	return out, md, liberrors.ErrSkipRequest(fmt.Sprintf("skip global: '%s'", middleware.GetServiceID(ctx)))
}

type traceRequest struct{}

func (traceRequest) ID() string {
	return "aws-nuke::traceRequest"
}

func (traceRequest) HandleFinalize(
	ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler,
) (
	middleware.FinalizeOutput, middleware.Metadata, error,
) {
	req, ok := in.Request.(*smithyhttp.Request)
	if ok {
		log.Tracef("sending AWS request:\n%s", DumpRequest(req.Request))
	}
	return next.HandleFinalize(ctx, in)
}

type traceResponse struct{}

func (traceResponse) ID() string {
	return "aws-nuke::traceResponse"
}

func (traceResponse) HandleDeserialize(
	ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler,
) (
	middleware.DeserializeOutput, middleware.Metadata, error,
) {
	out, md, err := next.HandleDeserialize(ctx, in)

	resp, ok := out.RawResponse.(*smithyhttp.Response)
	if ok {
		log.Tracef("received AWS response:\n%s", DumpResponse(resp.Response))
	}
	return out, md, err
}
