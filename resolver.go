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
	PREFIX     = "DNS:"
	DNS_PREFIX = "_nkn."
	TXT_PREFIX = "nkn"
)

var (
	// ErrInvalidDomain is returned when a string representing a domain name
	// is not actually a valid domain.
	ErrInvalidDomain = errors.New("not a valid domain name")

	// ErrInvalidNknlink is returned when the nknlink entry in a TXT record
	// does not follow the proper nknlink format ("nknlink=<path>")
	ErrInvalidNknlink = errors.New("not a valid nknlink entry")

	// ErrResolveFailed is returned when a resolution failed, most likely
	// due to a network error.
	ErrResolveFailed = errors.New("link resolution failed")
)

type Config struct {
	Prefix       string
	CacheTimeout time.Duration // seconds
	DialTimeout  int           // milliseconds
}

type Resolver struct {
	config *Config
	cache  *cache.Cache
}

var DefaultConfig = Config{
	Prefix:       PREFIX,
	CacheTimeout: cache.NoExpiration,
	DialTimeout:  5000,
}

func GetDefaultConfig() *Config {
	return &DefaultConfig
}

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

func ParseTXT(txt string) (string, error) {
	parts := strings.SplitN(txt, "=", 2)
	if len(parts) == 2 && parts[0] == TXT_PREFIX {
		return path.Clean(parts[1]), nil
	}

	return "", ErrInvalidNknlink
}

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

func (r *Resolver) Resolve(address string) (string, error) {
	if !strings.HasPrefix(address, r.config.Prefix) {
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
