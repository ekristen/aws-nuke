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

// GlobalServices is the set of AWS services that operate globally
// and should only be processed in the "global" pseudo-region.
// The keys are the service IDs as returned by middleware.GetServiceID(ctx).
var GlobalServices = map[string]struct{}{
	"CloudFront":               {},
	"IAM":                      {},
	"Route 53":                 {},
	"WAF":                      {}, // WAF Classic (global)
	"CloudFront KeyValueStore": {},
}

// IsGlobalService returns true if the service should only run in global region
func IsGlobalService(service string) bool {
	_, ok := GlobalServices[service]
	return ok
}

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

		// Copy() is shallow, so the APIOptions slice is shared with the root config.
		// Clone it before appending, otherwise each region's config can write into
		// the root's backing array whenever it has spare capacity.
		cfgCopy.APIOptions = cloneAPIOptions(root.APIOptions, 1)

		if global {
			cfgCopy.APIOptions = append(cfgCopy.APIOptions, func(stack *middleware.Stack) error {
				return stack.Initialize.Add(SkipGlobal{}, middleware.After)
			})
		} else {
			cfgCopy.APIOptions = append(cfgCopy.APIOptions, func(stack *middleware.Stack) error {
				return stack.Initialize.Add(SkipRegionalForGlobalService{}, middleware.After)
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

// SkipGlobal skips requests for non-global services when operating in the
// global pseudo-region. Global services (CloudFront, IAM, Route 53, etc.)
// are allowed through, while regional services are skipped.
type SkipGlobal struct{}

func (SkipGlobal) ID() string {
	return "aws-nuke::skipGlobal"
}

func (SkipGlobal) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, md middleware.Metadata, err error,
) {
	service := middleware.GetServiceID(ctx)

	if IsGlobalService(service) {
		// Global service in global region - allow
		return next.HandleInitialize(ctx, in)
	}

	// Non-global service in global region - skip
	return out, md, liberrors.ErrSkipRequest(
		fmt.Sprintf("service '%s' is not global, but the session is", service))
}

// SkipRegionalForGlobalService skips requests for global-only services
// when operating in a non-global region context. This ensures global
// services like CloudFront and IAM are only processed once in the
// "global" pseudo-region rather than in every regional scan.
type SkipRegionalForGlobalService struct{}

func (SkipRegionalForGlobalService) ID() string {
	return "aws-nuke::skipRegionalForGlobalService"
}

func (SkipRegionalForGlobalService) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, md middleware.Metadata, err error,
) {
	service := middleware.GetServiceID(ctx)

	if IsGlobalService(service) {
		return out, md, liberrors.ErrSkipRequest(
			fmt.Sprintf("service '%s' is global, but the session is not", service))
	}

	return next.HandleInitialize(ctx, in)
}

// CrossServiceConfig returns a copy of cfg with the region scanning guards removed.
//
// SkipGlobal and SkipRegionalForGlobalService exist so that a resource type is only
// scanned and deleted in the region that owns it: global services in the "global"
// pseudo-region, everything else in each real region. Because they are attached to the
// config, they also block every client built from that config, including ones a
// resource builds for a *different* service purely to support its own removal, such as
// a regional resource creating an IAM role so it can delete itself.
//
// Use this whenever building a client for a service other than the one the resource
// itself represents. Do not use it for the resource's own service; that would cause the
// resource to be processed in every region.
//
// The region is deliberately left untouched. Global services resolve to their global
// endpoint from any region in the partition, so overriding it would only break the
// non-default partitions.
func CrossServiceConfig(cfg *aws.Config) *aws.Config {
	c := cfg.Copy()

	// Copy() is shallow, so clone APIOptions rather than appending into the slice that
	// is still shared with cfg.
	c.APIOptions = cloneAPIOptions(cfg.APIOptions, 1)
	c.APIOptions = append(c.APIOptions, func(stack *middleware.Stack) error {
		// Only one of these is ever present, and neither is present when custom
		// endpoints are in play, so a missing middleware is not an error.
		_, _ = stack.Initialize.Remove(SkipGlobal{}.ID())
		_, _ = stack.Initialize.Remove(SkipRegionalForGlobalService{}.ID())
		return nil
	})

	return &c
}

// cloneAPIOptions copies a set of API options into a new slice with room for extra
// entries. aws.Config.Copy() is a shallow copy, so appending to the APIOptions of a
// copied config can otherwise write into the original config's backing array.
func cloneAPIOptions(opts []func(*middleware.Stack) error, extra int) []func(*middleware.Stack) error {
	cloned := make([]func(*middleware.Stack) error, len(opts), len(opts)+extra)
	copy(cloned, opts)
	return cloned
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
