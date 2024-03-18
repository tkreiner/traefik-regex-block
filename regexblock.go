package traefik_regex_block

import (
	"errors"
	"context"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"
        "fmt"

	"github.com/zerodha/logf"
)

var mylog = logf.New(logf.Opts{
                Level:                logf.DebugLevel,
                TimestampFormat:      time.RFC3339Nano,
        })

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
		EnableDebug: false,
	}
}

// RegexBlock is a Traefik plugin that blocks requests matching certain regex patterns.
type RegexBlock struct {
	next              http.Handler
        name              string
	regexPatterns     []*regexp.Regexp
	blockDuration     time.Duration
	whitelist         []*net.IPNet
	blockedIPs        map[string]time.Time
	mutex             sync.Mutex
}

// New creates a new instance of the RegexBlock.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	mylog = logf.New(logf.Opts{
                Level:                logf.DebugLevel,
                TimestampFormat:      time.RFC3339Nano,
                DefaultFields:        []any{"plugin", "traefik-regex-block", "pluginName", name},
        })

        mylog.Info("RegexBlock plugin is starting.")

	// Setup list of regex patterns
	regexPatterns := make([]*regexp.Regexp, 0)
	for _, pattern := range config.RegexPatterns {
		compiledRegex, err := regexp.Compile(pattern)
		if err != nil {
			mylog.Error(fmt.Sprintf("Regex pattern %s is invalid and will not be used.",pattern))
			continue
		}
		regexPatterns = append(regexPatterns, compiledRegex)
                mylog.Debug(fmt.Sprintf("Adding regex pattern %s",compiledRegex.String()))
	}
	if len(regexPatterns) == 0 {
		mylog.Error("There were no valid regex patterns. Plugin will not load.")
		return nil, errors.New("No valid regex patterns found.")
	}

	// Setup block duration
	blockDuration := time.Duration(config.BlockDurationMinutes) * time.Minute
        mylog.Info(fmt.Sprintf("Setting block duration as %d minutes.",config.BlockDurationMinutes))

	// Setup list of IP addresses to whitelist
	whitelist := make([]*net.IPNet, 0)
	for _, ip := range config.Whitelist {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			// Try parsing as single IP address
			ipAddr := net.ParseIP(ip)
			if ipAddr != nil {
				ipNet = &net.IPNet{IP: ipAddr, Mask: net.CIDRMask(32, 32)}
			} else {
				mylog.Error(fmt.Sprintf("Whitelist IP address %s is invalid and will not be used.",ip))
				continue
			}
		}
		whitelist = append(whitelist, ipNet)
                mylog.Debug(fmt.Sprintf("Adding whitelist IP %s",ip))
	}

	return &RegexBlock{
		next:              next,
		name:              name,
		regexPatterns:     regexPatterns,
		blockDuration:     blockDuration,
		whitelist:         whitelist,
		blockedIPs:        make(map[string]time.Time),
	}, nil
}

// ServeHTTP intercepts the request and blocks it if it matches any of the configured regex patterns.
func (p *RegexBlock) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
        mylog.Debug(fmt.Sprintf("Testing IP %s.",ip))

	// Check if IP is whitelisted
	if p.isWhitelisted(ip) {
		p.next.ServeHTTP(rw, req)
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if IP is blocked
	if blockTime, ok := p.blockedIPs[ip]; ok {
		if time.Since(blockTime) < p.blockDuration {
                        mylog.Debug(fmt.Sprintf("IP %s is still blocked.",ip))
			rw.WriteHeader(http.StatusForbidden)
			return
		} else {
                        mylog.Debug(fmt.Sprintf("Removing block for IP %s.",ip))
			delete(p.blockedIPs, ip) // Unblock the IP if the block time has expired
		}
	}

	// Check if the request matches any regex pattern
	for _, pattern := range p.regexPatterns {
		if pattern.MatchString(req.URL.Path) {
			// Block the IP for the specified duration
                        mylog.Info(fmt.Sprintf("Setting block for IP %s for requested path %s, based on regex of %s.",ip,req.URL.Path,pattern.String()))
			p.blockedIPs[ip] = time.Now()
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Allow the request to pass through
	p.next.ServeHTTP(rw, req)
}

// isWhitelisted checks if the IP address is whitelisted.
func (p *RegexBlock) isWhitelisted(ip string) bool {
        mylog.Debug("Checking if IP %s is in whitelist")
	addr := net.ParseIP(ip)
	if addr == nil {
		mylog.Debug(fmt.Sprintf("Could not parse request IP %s",ip))
		return false
	}

	for _, ipNet := range p.whitelist {
		if ipNet.Contains(addr) {
			mylog.Debug("IP %s is in whitelist")
			return true
		}
	}
	mylog.Debug("IP %s is not in whitelist")
	return false
}
