package traefik_regex_block

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// Config defines the configuration options for the plugin.
type Config struct {
	RegexPatterns        []string `json:"regexPatterns,omitempty"`
	BlockDurationMinutes int      `json:"blockDurationMinutes,omitempty"`
	Whitelist            []string `json:"whitelist,omitempty"`
	EnableDebug          bool     `json:"enableDebug,omitempty"`
}

// CreateConfig creates a default configuration for the plugin.
func CreateConfig() *Config {
	return &Config{
		BlockDurationMinutes: 60, // Default block duration: 1 hour
		EnableDebug:          false,
	}
}

// RegexBlock is a Traefik plugin that blocks requests matching certain regex patterns.
type RegexBlock struct {
	next          http.Handler
	name          string
	regexPatterns []*regexp.Regexp
	blockDuration int
	whitelist     []*net.IPNet
	blockedIPs    map[string]time.Time
	logger        *pluginLogger
	blockMgr      *BlockManager
	mutex         sync.Mutex
}

// New creates a new instance of the RegexBlock.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	logLevel := "info"
	if config.EnableDebug {
		logLevel = "debug"
	}
	logger := newPluginLogger(logLevel, name)
	logger.Info("RegexBlock plugin is starting.")

	// Setup list of regex patterns
	regexPatterns := make([]*regexp.Regexp, 0)
	for _, pattern := range config.RegexPatterns {
		compiledRegex, err := regexp.Compile(pattern)
		if err != nil {
			logger.Error(fmt.Sprintf("Regex pattern %s is invalid and will not be used.", pattern))
			continue
		}
		regexPatterns = append(regexPatterns, compiledRegex)
		logger.Debug(fmt.Sprintf("Adding regex pattern %s", compiledRegex.String()))
	}
	if len(regexPatterns) == 0 {
		logger.Error("There were no valid regex patterns. Plugin will not load.")
		return nil, errors.New("No valid regex patterns found.")
	}

	// Setup block duration
	blockDuration := config.BlockDurationMinutes
	logger.Info(fmt.Sprintf("Setting block duration as %d minutes.", blockDuration))

	// Setup list of IP addresses to whitelist
	whitelist := make([]*net.IPNet, 0)
	for _, ip := range config.Whitelist {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			// Try parsing as single IP address
			ipAddr := net.ParseIP(ip)
			if ipAddr != nil {
				// Determine if IPv4 or IPv6 and use appropriate mask
				if ipAddr.To4() != nil {
					ipNet = &net.IPNet{IP: ipAddr, Mask: net.CIDRMask(32, 32)}
				} else {
					ipNet = &net.IPNet{IP: ipAddr, Mask: net.CIDRMask(128, 128)}
				}
			} else {
				logger.Error(fmt.Sprintf("Whitelist IP address %s is invalid and will not be used.", ip))
				continue
			}
		}
		whitelist = append(whitelist, ipNet)
		logger.Debug(fmt.Sprintf("Adding whitelist IP %s", ip))
	}

	// Setup a manager for the block list. Currently only
	// supports an array. Future plans to support Redis and/or MySQL
	blockMgr := ArrayBlockManager()

	return &RegexBlock{
		next:          next,
		name:          name,
		regexPatterns: regexPatterns,
		blockDuration: blockDuration,
		whitelist:     whitelist,
		blockedIPs:    make(map[string]time.Time),
		logger:        logger,
		blockMgr:      blockMgr,
	}, nil
}

// ServeHTTP intercepts the request and blocks it if it matches any of the configured regex patterns.
func (p *RegexBlock) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port, try parsing as-is
		ip = req.RemoteAddr
		p.logger.Debug(fmt.Sprintf("Could not split host and port from RemoteAddr %s, using as-is: %v", req.RemoteAddr, err))
	}
	ipNet := net.ParseIP(ip)
	if ipNet == nil {
		p.logger.Error(fmt.Sprintf("Could not parse IP address from RemoteAddr %s, allowing request", ip))
		p.next.ServeHTTP(rw, req)
		return
	}
	p.logger.Debug(fmt.Sprintf("Testing IP %s.", ip))

	// Check if IP is whitelisted
	if p.isWhitelisted(ip) {
		p.next.ServeHTTP(rw, req)
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if IP is blocked
	if p.blockMgr.IsBlocked(ipNet) {
		p.logger.Debug(fmt.Sprintf("IP %s is still blocked.", ip))
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if the request matches any regex pattern
	for _, pattern := range p.regexPatterns {
		if pattern.MatchString(req.URL.Path) {
			// Block the IP for the specified duration
			p.logger.Info(fmt.Sprintf("Setting block for IP %s for requested path %s, based on regex of %s.", ip, req.URL.Path, pattern.String()))
			p.blockMgr.Block(ipNet, p.blockDuration)
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Allow the request to pass through
	p.next.ServeHTTP(rw, req)
}

// isWhitelisted checks if the IP address is whitelisted.
func (p *RegexBlock) isWhitelisted(ip string) bool {
	p.logger.Debug(fmt.Sprintf("Checking if IP %s is in whitelist", ip))
	addr := net.ParseIP(ip)
	if addr == nil {
		p.logger.Debug(fmt.Sprintf("Could not parse request IP %s", ip))
		return false
	}

	for _, ipNet := range p.whitelist {
		if ipNet.Contains(addr) {
			p.logger.Debug(fmt.Sprintf("IP %s is in whitelist", ip))
			return true
		}
	}
	p.logger.Debug(fmt.Sprintf("IP %s is not in whitelist", ip))
	return false
}
