package nuke

import (
	"context"
	"fmt"
	"sync"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session" //nolint:staticcheck

	liberrors "github.com/ekristen/libnuke/pkg/errors"
)

// SessionFactory support for custom endpoints
type SessionFactory func(regionName, svcType string) (*session.Session, error)

// ConfigFactory is the SDK v2 equivalent to SessionFactory.
type ConfigFactory func(ctx context.Context, regionName, svcType string) (*awsv2.Config, error)

// ResourceTypeResolver returns the service type from the resourceType
type ResourceTypeResolver func(regionName, resourceType string) string

// Region is an AWS Region with an attached SessionFactory
type Region struct {
	Name            string
	NewSession      SessionFactory // SDK v1
	NewConfig       ConfigFactory  // SDK v2
	ResTypeResolver ResourceTypeResolver

	cache    map[string]*session.Session
	cfgCache map[string]*awsv2.Config
	lock     *sync.RWMutex
}

// NewRegion creates a new Region and returns it.
func NewRegion(name string, typeResolver ResourceTypeResolver, sessionFactory SessionFactory, cfgFactory ConfigFactory) *Region {
	return &Region{
		Name:            name,
		NewSession:      sessionFactory,
		NewConfig:       cfgFactory,
		ResTypeResolver: typeResolver,
		lock:            &sync.RWMutex{},
		cache:           make(map[string]*session.Session),
		cfgCache:        make(map[string]*awsv2.Config),
	}
}

// Session returns a session for a given resource type for the region it's associated to.
func (region *Region) Session(resourceType string) (*session.Session, error) {
	svcType := region.ResTypeResolver(region.Name, resourceType)
	if svcType == "" {
		return nil, liberrors.ErrSkipRequest(fmt.Sprintf(
			"No service available in region '%s' to handle '%s'",
			region.Name, resourceType))
	}

	// Need to read
	region.lock.RLock()
	sess := region.cache[svcType]
	region.lock.RUnlock()
	if sess != nil {
		return sess, nil
	}

	// Need to write:
	region.lock.Lock()
	sess, err := region.NewSession(region.Name, svcType)
	if err != nil {
		region.lock.Unlock()
		return nil, err
	}
	region.cache[svcType] = sess
	region.lock.Unlock()
	return sess, nil
}

// Config returns an SDK v2 config for a given resource type for the region
// it's associated to.
func (region *Region) Config(resourceType string) (*awsv2.Config, error) {
	svcType := region.ResTypeResolver(region.Name, resourceType)
	if svcType == "" {
		return nil, liberrors.ErrSkipRequest(fmt.Sprintf(
			"No service available in region '%s' to handle '%s'",
			region.Name, resourceType))
	}

	// Need to read
	region.lock.RLock()
	cfg := region.cfgCache[svcType]
	region.lock.RUnlock()
	if cfg != nil {
		return cfg, nil
	}

	// Need to write:
	region.lock.Lock()
	cfg, err := region.NewConfig(context.TODO(), region.Name, svcType)
	if err != nil {
		region.lock.Unlock()
		return nil, err
	}
	region.cfgCache[svcType] = cfg
	region.lock.Unlock()
	return cfg, nil
}
