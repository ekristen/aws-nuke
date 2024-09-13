package config

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	libconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"
)

// TestConfig_NewMissingFile tests the extended config loads functionality when the file is missing.
func TestConfig_NewMissingFile(t *testing.T) {
	_, err := New(libconfig.Options{
		Path: "testdata/missing.yaml",
	})
	assert.Error(t, err)
}

// TestConfig_MissingFile tests the extended config loads functionality when the file is missing.
// We do not use the New here because that would test the libnuke config loading functionality.
func TestConfig_MissingFile(t *testing.T) {
	cfg := Config{}
	err := cfg.Load("testdata/missing.yaml")
	assert.Error(t, err)
}

// TestConfig_InvalidFile tests the extended config loads functionality when the file is invalid.
// We do not use the New here because that would test the libnuke config loading functionality.
func TestConfig_InvalidFile(t *testing.T) {
	cfg := Config{}
	err := cfg.Load("testdata/invalid.yaml")
	assert.Error(t, err)
}

func TestConfig_LoadExample(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	entry := logrus.WithField("test", true)

	config, err := New(libconfig.Options{
		Path: "testdata/example.yaml",
		Log:  entry,
	})
	if err != nil {
		t.Fatal(err)
	}

	expect := Config{
		Config: &libconfig.Config{
			Blocklist: []string{"1234567890"},
			Regions:   []string{"eu-west-1", "stratoscale"},
			Accounts: map[string]*libconfig.Account{
				"555133742": {
					Presets: []string{"terraform"},
					Filters: filter.Filters{
						"IAMRole": {
							filter.NewExactFilter("uber.admin"),
						},
						"IAMRolePolicyAttachment": {
							filter.NewExactFilter("uber.admin -> AdministratorAccess"),
						},
					},
					ResourceTypes: libconfig.ResourceTypes{
						Targets: types.Collection{"S3Bucket"},
					},
				},
			},
			ResourceTypes: libconfig.ResourceTypes{
				Targets:  types.Collection{"DynamoDBTable", "S3Bucket", "S3Object"},
				Excludes: types.Collection{"IAMRole"},
			},
			Presets: map[string]libconfig.Preset{
				"terraform": {
					Filters: filter.Filters{
						"S3Bucket": {
							filter.Filter{
								Type:   filter.Glob,
								Value:  "my-statebucket-*",
								Values: []string{},
							},
						},
					},
				},
			},
			Settings:     &settings.Settings{},
			Deprecations: make(map[string]string),
			Log:          entry,
		},
		CustomEndpoints: []*CustomRegion{
			{
				Region:                "stratoscale",
				TLSInsecureSkipVerify: true,
				Services: CustomServices{
					&CustomService{
						Service: "ec2",
						URL:     "https://stratoscale.cloud.internal/api/v2/aws/ec2",
					},
					&CustomService{
						Service:               "s3",
						URL:                   "https://stratoscale.cloud.internal:1060",
						TLSInsecureSkipVerify: true,
					},
				},
			},
		},
	}

	assert.Equal(t, expect, *config)
}

func TestConfig_ResolveDeprecations(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	entry := logrus.WithField("test", true)

	config := Config{
		Config: &libconfig.Config{
			Blocklist: []string{"1234567890"},
			Regions:   []string{"eu-west-1"},
			Accounts: map[string]*libconfig.Account{
				"555133742": {
					Filters: filter.Filters{
						"IamRole": {
							filter.NewExactFilter("uber.admin"),
							filter.NewExactFilter("foo.bar"),
						},
						"IAMRolePolicyAttachment": {
							filter.NewExactFilter("uber.admin -> AdministratorAccess"),
						},
					},
				},
				"2345678901": {
					Filters: filter.Filters{
						"ECRrepository": {
							filter.NewExactFilter("foo:bar"),
							filter.NewExactFilter("bar:foo"),
						},
						"IAMRolePolicyAttachment": {
							filter.NewExactFilter("uber.admin -> AdministratorAccess"),
						},
					},
				},
			},
			Settings: &settings.Settings{},
			Deprecations: map[string]string{
				"ECRrepository": "ECRRepository",
				"IamRole":       "IAMRole",
			},
			Log: entry,
		},
	}

	expect := map[string]*libconfig.Account{
		"555133742": {
			Filters: filter.Filters{
				"IAMRole": {
					filter.NewExactFilter("uber.admin"),
					filter.NewExactFilter("foo.bar"),
				},
				"IAMRolePolicyAttachment": {
					filter.NewExactFilter("uber.admin -> AdministratorAccess"),
				},
			},
		},
		"2345678901": {
			Filters: filter.Filters{
				"ECRRepository": {
					filter.NewExactFilter("foo:bar"),
					filter.NewExactFilter("bar:foo"),
				},
				"IAMRolePolicyAttachment": {
					filter.NewExactFilter("uber.admin -> AdministratorAccess"),
				},
			},
		},
	}

	err := config.ResolveDeprecations()
	assert.NoError(t, err)

	assert.Equal(t, expect, config.Accounts)

	invalidConfig := Config{
		Config: &libconfig.Config{
			Blocklist: []string{"1234567890"},
			Regions:   []string{"eu-west-1"},
			Accounts: map[string]*libconfig.Account{
				"555133742": {
					Filters: filter.Filters{
						"IamUserAccessKeys": {
							filter.NewExactFilter("X")},
						"IAMUserAccessKey": {
							filter.NewExactFilter("Y")},
					},
				},
			},
			Deprecations: map[string]string{
				"IamUserAccessKeys": "IAMUserAccessKey",
			},
			Log: entry,
		},
	}

	err = invalidConfig.ResolveDeprecations()
	if err == nil || !strings.Contains(err.Error(), "using deprecated resource type and replacement") {
		t.Fatal("invalid config did not cause correct error")
	}
}

func TestConfig_Validation(t *testing.T) {
	cases := []struct {
		ID             string
		Aliases        []string
		ShouldFail     bool
		SkipAliasCheck bool
		Config         string
	}{
		{ID: "555133742", Aliases: []string{"staging"}, ShouldFail: false, Config: "testdata/example.yaml"},
		{ID: "1234567890", Aliases: []string{"staging"}, ShouldFail: true, Config: "testdata/example.yaml"},
		{ID: "1111111111", Aliases: []string{"staging"}, ShouldFail: true, Config: "testdata/example.yaml"},
		{ID: "555133742", Aliases: []string{"production"}, ShouldFail: true, Config: "testdata/example.yaml"},
		{ID: "555133742", Aliases: []string{}, ShouldFail: true, Config: "testdata/example.yaml"},
		{ID: "555133742", Aliases: []string{"staging", "prod"}, ShouldFail: true, Config: "testdata/example.yaml"},
		{ID: "555133742", Aliases: []string{}, ShouldFail: true, SkipAliasCheck: true, Config: "testdata/example.yaml"},
		{ID: "123654654", Aliases: []string{}, ShouldFail: false, SkipAliasCheck: true, Config: "testdata/bypass-alias.yaml"},
	}

	for i, tc := range cases {
		name := fmt.Sprintf("%d_%s/%v/%t", i, tc.ID, tc.Aliases, tc.ShouldFail)
		t.Run(name, func(t *testing.T) {
			config, err := New(libconfig.Options{
				Path: tc.Config,
			})
			assert.NoError(t, err)

			vErr := config.ValidateAccount(tc.ID, tc.Aliases, tc.SkipAliasCheck)
			if tc.ShouldFail {
				assert.Error(t, vErr)
			} else {
				assert.NoError(t, vErr)
			}
		})
	}
}

func TestConfig_DeprecatedKeys(t *testing.T) {
	config, err := New(libconfig.Options{
		Path: "testdata/deprecated-keys-config.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !config.InBlocklist("1234567890") {
		t.Errorf("Loading the config did not resolve the deprecated key 'account-blacklist' correctly")
	}
}

func TestConfig_FilterMerge(t *testing.T) {
	config, err := New(libconfig.Options{
		Path: "testdata/example.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}

	filters, err := config.Filters("555133742")
	if err != nil {
		t.Fatal(err)
	}

	expect := filter.Filters{
		"S3Bucket": []filter.Filter{
			{
				Type: "glob", Value: "my-statebucket-*", Values: []string{},
			},
		},
		"IAMRole": []filter.Filter{
			{
				Type:  "exact",
				Value: "uber.admin",
			},
		},
		"IAMRolePolicyAttachment": []filter.Filter{
			{
				Type:  "exact",
				Value: "uber.admin -> AdministratorAccess",
			},
		},
	}

	if !reflect.DeepEqual(filters, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", filters)
		t.Errorf("  Expected: %#v", expect)
	}
}

func TestConfig_GetCustomRegion(t *testing.T) {
	config, err := New(libconfig.Options{
		Path: "testdata/example.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}
	stratoscale := config.CustomEndpoints.GetRegion("stratoscale")
	if stratoscale == nil {
		t.Fatal("Expected to find a set of custom endpoints for region10")
	}
	euWest1 := config.CustomEndpoints.GetRegion("eu-west-1")
	if euWest1 != nil {
		t.Fatal("Expected to euWest1 without a set of custom endpoints")
	}

	assert.Equal(t, "https://stratoscale.cloud.internal/api/v2/aws/ec2",
		config.CustomEndpoints.GetURL("stratoscale", "ec2"))

	assert.Equal(t, "", config.CustomEndpoints.GetURL("invalid", "rds"))
	assert.Equal(t, "", config.CustomEndpoints.GetURL("stratoscale", "rds"))

	t.Run("TestGetService", func(t *testing.T) {
		ec2Service := stratoscale.Services.GetService("ec2")
		if ec2Service == nil {
			t.Fatal("Expected to find a custom ec2 service for region10")
		}
		rdsService := stratoscale.Services.GetService("rds")
		if rdsService != nil {
			t.Fatal("Expected to not find a custom rds service for region10")
		}
	})
}

func TestConfig_DeprecatedFeatureFlags(t *testing.T) {
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/config/config.go") {
				return
			}

			if e.Caller.Line == 235 {
				assert.Equal(t, "deprecated configuration key 'feature-flags' - please use 'settings' instead", e.Message)
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	opts := libconfig.Options{
		Path: "testdata/deprecated-feature-flags.yaml",
	}

	c, err := New(opts)

	assert.NoError(t, err)
	assert.NotNil(t, c)

	ec2InstanceSettings := c.Settings.Get("EC2Instance")
	assert.NotNil(t, ec2InstanceSettings)
	assert.Equal(t, true, ec2InstanceSettings.Get("DisableDeletionProtection"))
	assert.Equal(t, true, ec2InstanceSettings.Get("DisableStopProtection"))

	rdsInstanceSettings := c.Settings.Get("RDSInstance")
	assert.NotNil(t, rdsInstanceSettings)
	assert.Equal(t, true, rdsInstanceSettings.Get("DisableDeletionProtection"))

	cloudformationStackSettings := c.Settings.Get("CloudFormationStack")
	assert.NotNil(t, cloudformationStackSettings)
	assert.Equal(t, true, cloudformationStackSettings.Get("DisableDeletionProtection"))
}

func TestConfig_ValidateAccount_Blocklist(t *testing.T) {
	config, err := New(libconfig.Options{
		Path: "testdata/example.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Add an account to the blocklist
	config.Blocklist = append(config.Blocklist, "1234567890")
	config.BlocklistAliasKeywords = append(config.BlocklistAliasKeywords, "alpha-tango")

	// Test cases
	cases := []struct {
		ID         string
		Aliases    []string
		ShouldFail bool
	}{
		{
			// Should fail due to blocklist
			ID: "1234567890",
			Aliases: []string{
				"sandbox",
			},
			ShouldFail: true,
		},
		{
			// Allowed account
			ID: "555133742",
			Aliases: []string{
				"sandbox2",
			},
			ShouldFail: false,
		},
		{
			// Allowed account but blocked by keyword
			ID: "555133742",
			Aliases: []string{
				"alpha-tango-sandbox",
			},
			ShouldFail: true,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("AccountID_%s", tc.ID), func(t *testing.T) {
			err := config.ValidateAccount(tc.ID, tc.Aliases, false)
			if tc.ShouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
