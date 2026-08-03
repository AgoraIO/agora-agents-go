package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Area represents the global regions where the Open API gateway endpoint is located
type Area int

const (
	AreaUnknown Area = iota
	// AreaUS represents the western and eastern regions of the United States
	AreaUS
	// AreaEU represents the western and central regions of Europe
	AreaEU
	// AreaAP represents the southeastern and northeastern regions of Asia-Pacific
	AreaAP
	// AreaCN represents the eastern and northern regions of Chinese mainland
	AreaCN
)

const (
	ChineseMainlandMajorDomain = "sd-rtn.com"
	OverseaMajorDomain         = "agora.io"
)

const GlobalDomainPrefix = "api"

const (
	GlobalConvoAIAPIPathSuffix = "/api/conversational-ai-agent"
	CNConvoAICNAPIPathSuffix   = "/cn/api/conversational-ai-agent"
)

const apiBaseURLEnv = "AGORA_AGENTS_API_BASE_URL"

const (
	USWestRegionDomainPrefix = "api-us-west-1"
	USEastRegionDomainPrefix = "api-us-east-1"
)

const (
	APSoutheastRegionDomainPrefix = "api-ap-southeast-1"
	APNortheastRegionDomainPrefix = "api-ap-northeast-1"
)

const (
	EUWestRegionDomainPrefix    = "api-eu-west-1"
	EUCentralRegionDomainPrefix = "api-eu-central-1"
)

const (
	CNEastRegionDomainPrefix  = "api-cn-east-1"
	CNNorthRegionDomainPrefix = "api-cn-north-1"
)

// Domain contains the regional prefixes and domain suffixes for an area
type Domain struct {
	RegionDomainPrefixes []string
	MajorDomainSuffixes  []string
}

// RegionDomain maps areas to their domain configurations
var RegionDomain = map[Area]Domain{
	AreaUS: {
		RegionDomainPrefixes: []string{
			USWestRegionDomainPrefix,
			USEastRegionDomainPrefix,
		},
		MajorDomainSuffixes: []string{
			OverseaMajorDomain,
			ChineseMainlandMajorDomain,
		},
	},
	AreaEU: {
		RegionDomainPrefixes: []string{
			EUWestRegionDomainPrefix,
			EUCentralRegionDomainPrefix,
		},
		MajorDomainSuffixes: []string{
			OverseaMajorDomain,
			ChineseMainlandMajorDomain,
		},
	},
	AreaAP: {
		RegionDomainPrefixes: []string{
			APSoutheastRegionDomainPrefix,
			APNortheastRegionDomainPrefix,
		},
		MajorDomainSuffixes: []string{
			OverseaMajorDomain,
			ChineseMainlandMajorDomain,
		},
	},
	AreaCN: {
		RegionDomainPrefixes: []string{
			CNEastRegionDomainPrefix,
			CNNorthRegionDomainPrefix,
		},
		MajorDomainSuffixes: []string{
			ChineseMainlandMajorDomain,
			OverseaMajorDomain,
		},
	},
}

// Resolver is an interface for resolving the best domain
type Resolver interface {
	Resolve(ctx context.Context, domains []string, regionPrefix string) (string, error)
}

// ResolverFunc is a function type that implements Resolver
type ResolverFunc func(ctx context.Context, domains []string, regionPrefix string) (string, error)

func (r ResolverFunc) Resolve(ctx context.Context, domains []string, regionPrefix string) (string, error) {
	return r(ctx, domains, regionPrefix)
}

// resolverImpl is the default DNS-based resolver implementation
type resolverImpl struct{}

func newResolverImpl() *resolverImpl {
	return &resolverImpl{}
}

func (r *resolverImpl) Resolve(ctx context.Context, domains []string, regionPrefix string) (string, error) {
	var wg sync.WaitGroup

	done := make(chan struct{}, 1)
	res := make(chan string, len(domains))
	for _, d := range domains {
		wg.Add(1)

		go func(domain string, regionPrefix string) {
			defer wg.Done()
			url := regionPrefix + "." + domain
			_, err := net.LookupHost(url)
			if err == nil {
				res <- domain
			}
		}(d, regionPrefix)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case domain := <-res:
		return domain, nil
	case <-done:
	}
	return "", errors.New("query all dns failed")
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Pool manages a pool of regional URLs with automatic cycling and domain selection
type Pool struct {
	domainArea            Area
	domainSuffixes        []string
	currentDomain         string
	regionPrefixes        []string
	currentRegionPrefixes []string
	configuredBaseURL     string
	locker                *sync.Mutex

	resolver   Resolver
	lastUpdate time.Time
}

const updateDuration = 30 * time.Second

// NewPool creates a new domain pool for the specified area
func NewPool(domainArea Area) (*Pool, error) {
	if _, ok := RegionDomain[domainArea]; !ok {
		return nil, errors.New("invalid domain area")
	}

	configuredBaseURL, err := configuredBaseURL(domainArea)
	if err != nil {
		return nil, err
	}

	p := &Pool{
		domainArea:        domainArea,
		domainSuffixes:    RegionDomain[domainArea].MajorDomainSuffixes,
		configuredBaseURL: configuredBaseURL,
		resolver:          newResolverImpl(),
		locker:            &sync.Mutex{},
	}

	p.regionPrefixes = append(p.regionPrefixes, RegionDomain[domainArea].RegionDomainPrefixes...)
	p.currentRegionPrefixes = p.regionPrefixes
	p.currentDomain = p.domainSuffixes[0]

	return p, nil
}

func (p *Pool) domainNeedUpdate() bool {
	return time.Since(p.lastUpdate) > updateDuration
}

// SelectBestDomain uses DNS resolution to select the best available domain
func (p *Pool) SelectBestDomain(ctx context.Context) error {
	if p.configuredBaseURL != "" {
		return nil
	}

	if !p.domainNeedUpdate() {
		return nil
	}

	p.locker.Lock()
	defer p.locker.Unlock()

	if p.domainNeedUpdate() {
		domain, err := p.resolver.Resolve(ctx, p.domainSuffixes, p.currentRegionPrefixes[0])
		if err != nil {
			return err
		}
		p.selectDomain(domain)
	}
	return nil
}

// NextRegion cycles to the next region prefix in the pool
func (p *Pool) NextRegion() {
	if p.configuredBaseURL != "" {
		return
	}

	p.locker.Lock()
	defer p.locker.Unlock()

	p.currentRegionPrefixes = p.currentRegionPrefixes[1:]
	if len(p.currentRegionPrefixes) == 0 {
		p.currentRegionPrefixes = p.regionPrefixes
	}
}

func (p *Pool) selectDomain(domain string) {
	if contains(p.domainSuffixes, domain) {
		p.currentDomain = domain
		p.lastUpdate = time.Now()
	}
}

// GetCurrentURL returns the current URL based on the selected region and domain
func (p *Pool) GetCurrentURL() string {
	p.locker.Lock()
	defer p.locker.Unlock()

	if p.configuredBaseURL != "" {
		return p.configuredBaseURL
	}

	currentRegion := p.currentRegionPrefixes[0]
	currentDomain := p.currentDomain
	return fmt.Sprintf("https://%s.%s%s", currentRegion, currentDomain, convoAIAPIPathSuffix(p.domainArea))
}

func configuredBaseURL(area Area) (string, error) {
	rawBaseURL := strings.TrimSpace(os.Getenv(apiBaseURLEnv))
	if rawBaseURL == "" {
		return "", nil
	}

	parsedURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid %s: %w", apiBaseURLEnv, err)
	}

	return parsedURL.JoinPath(convoAIAPIPathSuffix(area)).String(), nil
}

func convoAIAPIPathSuffix(area Area) string {
	if area == AreaCN {
		return CNConvoAICNAPIPathSuffix
	}
	return GlobalConvoAIAPIPathSuffix
}

// AreaRequestOption implements the RequestOption interface for area-based URL selection
type AreaRequestOption struct {
	Pool *Pool
}

func (o *AreaRequestOption) applyRequestOptions(opts *RequestOptions) {
	if o.Pool != nil {
		opts.BaseURL = o.Pool.GetCurrentURL()
	}
}

// NewAreaRequestOption creates a new AreaRequestOption with a pool for the specified area
func NewAreaRequestOption(area Area) *AreaRequestOption {
	pool, err := NewPool(area)
	if err != nil {
		panic(err)
	}
	return &AreaRequestOption{Pool: pool}
}
