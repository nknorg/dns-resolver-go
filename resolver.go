package dnsresolver

import (
	"context"
	"errors"
	"github.com/imdario/mergo"
	isdomain "github.com/jbenet/go-is-domain"
	"github.com/patrickmn/go-cache"
	"net"
	"path"

	"strings"
	"time"
)

const (
	// PREFIX Protocol prefix
	PREFIX = "DNS:"

	// DNS_PREFIX domain prefix
	DNS_PREFIX = "_nkn."

	// TXT_PREFIX  TXT record parameter name
	TXT_PREFIX = "nkn"
)

var (
	// ErrInvalidDomain is returned when a string representing a domain name
	// is not actually a valid domain.
	ErrInvalidDomain = errors.New("not a valid domain name")

	// ErrInvalidRecord is returned when the nkn entry in a TXT record
	// does not follow the proper nkn format ("nkn=<path>")
	ErrInvalidRecord = errors.New("not a valid record entry")

	// ErrResolveFailed is returned when a resolution failed, most likely
	// due to a network error.
	ErrResolveFailed = errors.New("record resolution failed")
)

// Config is the Resolver configuration.
type Config struct {
	Prefix       string
	CacheTimeout time.Duration // seconds
	DialTimeout  int           // milliseconds
	DnsServer    string
}

// Resolver implement ETH resolver.
type Resolver struct {
	config *Config
	cache  *cache.Cache
}

// DefaultConfig is the default Resolver config.
var DefaultConfig = Config{
	Prefix:       PREFIX,
	CacheTimeout: cache.NoExpiration,
	DialTimeout:  5000,
	DnsServer:    "",
}

// GetDefaultConfig returns the default Resolver config with nil pointer
// fields set to default.
func GetDefaultConfig() *Config {
	return &DefaultConfig
}

// MergeConfig merges a given Resolver config with the default Resolver config
// recursively. Any non zero value fields will override the default config.
func MergeConfig(config *Config) (*Config, error) {
	merged := GetDefaultConfig()
	if config != nil {
		err := mergo.Merge(merged, config, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
	}

	return merged, nil
}

// ParseTXT parses a TXT record value.
func ParseTXT(txt string) (string, error) {
	parts := strings.SplitN(txt, "=", 2)
	if len(parts) == 2 && parts[0] == TXT_PREFIX {
		return path.Clean(parts[1]), nil
	}

	return "", ErrInvalidRecord
}

// NewResolver creates a Resolver. If config is nil, the default Resolver config will be used.
func NewResolver(config *Config) (*Resolver, error) {
	config, err := MergeConfig(config)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		config: config,
		cache:  cache.New(config.CacheTimeout*time.Second, 60*time.Second),
	}, nil
}

// Resolve resolves the address and returns the mapping address.
func (r *Resolver) Resolve(address string) (string, error) {
	if !strings.HasPrefix(strings.ToUpper(address), r.config.Prefix) {
		return "", nil
	}

	address = address[len(r.config.Prefix):]
	addr := address
	addrCache, ok := r.cache.Get(address)
	if ok {
		addr = addrCache.(string)
		return addr, nil
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if r.config.DialTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.config.DialTimeout)*time.Millisecond)
		defer cancel()
	}

	if !isdomain.IsDomain(address) {
		return "", ErrInvalidDomain
	}

	dnsResolver := &net.Resolver{}
	if len(r.config.DnsServer) > 0 {
		dnsResolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			d := &net.Dialer{}
			conn, err := d.DialContext(ctx, network, r.config.DnsServer)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}
	}

	txt, err := dnsResolver.LookupTXT(ctx, DNS_PREFIX+address)
	if err != nil {
		return "", err
	}

	for _, t := range txt {
		p, err := ParseTXT(t)
		if err == nil {
			r.cache.Set(address, p, cache.DefaultExpiration)
			return p, nil
		}
	}

	return "", ErrResolveFailed
}
