// Package config owns validated startup configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sojournerdev/anya/internal/provider"
)

const (
	refreshIntervalEnvVar   = "AGENT_REFRESH_INTERVAL"
	refreshTimeoutEnvVar    = "AGENT_REFRESH_TIMEOUT"
	maxStateAgeEnvVar       = "AGENT_MAX_STATE_AGE"
	httpListenAddressEnvVar = "AGENT_HTTP_ADDRESS"
	providerEnvVar          = "AGENT_PROVIDER"
)

const (
	// DefaultRefreshInterval is used when AGENT_REFRESH_INTERVAL is unset.
	DefaultRefreshInterval time.Duration = 5 * time.Second
	minimumRefreshInterval time.Duration = time.Second
	maximumRefreshInterval time.Duration = 5 * time.Minute

	// DefaultRefreshTimeout is used when AGENT_REFRESH_TIMEOUT is unset.
	DefaultRefreshTimeout time.Duration = 2 * time.Second
	minimumRefreshTimeout time.Duration = 100 * time.Millisecond
	maximumRefreshTimeout time.Duration = time.Minute

	// DefaultMaxStateAge is used when AGENT_MAX_STATE_AGE is unset.
	DefaultMaxStateAge time.Duration = 30 * time.Second
	minimumMaxStateAge time.Duration = time.Second
	maximumMaxStateAge time.Duration = time.Hour

	// DefaultHTTPListenAddress is used when AGENT_HTTP_ADDRESS is unset.
	DefaultHTTPListenAddress = ":8080"
)

// Config contains the complete validated startup configuration.
type Config struct {
	Refresh  RefreshConfig
	HTTP     HTTPConfig
	Provider provider.Kind
}

// RefreshConfig contains refresh scheduling and freshness settings.
type RefreshConfig struct {
	Interval time.Duration
	Timeout  time.Duration
	MaxAge   time.Duration
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	ListenAddress string
}

// Load reads, parses, defaults, and validates startup configuration.
func Load() (Config, error) {
	refresh, err := loadRefreshConfig()
	if err != nil {
		return Config{}, err
	}

	httpConfig, err := loadHTTPConfig()
	if err != nil {
		return Config{}, err
	}

	providerKind, err := loadProviderKind()
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Refresh:  refresh,
		HTTP:     httpConfig,
		Provider: providerKind,
	}
	if err := config.validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) validate() error {
	if err := c.Refresh.validate(); err != nil {
		return fmt.Errorf("config: refresh: %w", err)
	}
	if err := c.HTTP.validate(); err != nil {
		return fmt.Errorf("config: HTTP: %w", err)
	}
	if c.Provider == "" {
		return errors.New("config: provider is required")
	}
	return nil
}

func (c RefreshConfig) validate() error {
	switch {
	case c.Interval <= 0:
		return errors.New("interval must be positive")
	case c.Timeout <= 0:
		return errors.New("timeout must be positive")
	case c.MaxAge <= 0:
		return errors.New("max state age must be positive")
	case c.Timeout > c.Interval:
		return fmt.Errorf(
			"timeout %s must not exceed interval %s",
			c.Timeout,
			c.Interval,
		)
	case c.MaxAge < c.Interval:
		return fmt.Errorf(
			"max state age %s must be at least interval %s",
			c.MaxAge,
			c.Interval,
		)
	default:
		return nil
	}
}

func (c HTTPConfig) validate() error {
	if strings.TrimSpace(c.ListenAddress) == "" {
		return errors.New("listen address is required")
	}
	return nil
}

func loadRefreshConfig() (RefreshConfig, error) {
	interval, err := loadDurationSetting(
		refreshIntervalEnvVar,
		DefaultRefreshInterval,
		minimumRefreshInterval,
		maximumRefreshInterval,
	)
	if err != nil {
		return RefreshConfig{}, err
	}

	timeout, err := loadDurationSetting(
		refreshTimeoutEnvVar,
		DefaultRefreshTimeout,
		minimumRefreshTimeout,
		maximumRefreshTimeout,
	)
	if err != nil {
		return RefreshConfig{}, err
	}

	maxAge, err := loadDurationSetting(
		maxStateAgeEnvVar,
		DefaultMaxStateAge,
		minimumMaxStateAge,
		maximumMaxStateAge,
	)
	if err != nil {
		return RefreshConfig{}, err
	}

	config := RefreshConfig{Interval: interval, Timeout: timeout, MaxAge: maxAge}
	if err := config.validate(); err != nil {
		return RefreshConfig{}, err
	}
	return config, nil
}

func loadDurationSetting(
	envVar string,
	defaultDuration time.Duration,
	minimumDuration time.Duration,
	maximumDuration time.Duration,
) (time.Duration, error) {
	rawValue, isConfigured := os.LookupEnv(envVar)
	if !isConfigured || rawValue == "" {
		return defaultDuration, nil
	}

	duration, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf(
			`config: %s must be a Go duration such as %q: %w`,
			envVar,
			defaultDuration.String(),
			err,
		)
	}
	if duration < minimumDuration || duration > maximumDuration {
		return 0, fmt.Errorf(
			`config: %s %q is outside the allowed range; use a value between %s and %s`,
			envVar,
			rawValue,
			minimumDuration,
			maximumDuration,
		)
	}
	return duration, nil
}

func loadHTTPConfig() (HTTPConfig, error) {
	listenAddress, isConfigured := os.LookupEnv(httpListenAddressEnvVar)
	if !isConfigured || strings.TrimSpace(listenAddress) == "" {
		return HTTPConfig{ListenAddress: DefaultHTTPListenAddress}, nil
	}

	host, portText, err := net.SplitHostPort(listenAddress)
	if err != nil {
		return HTTPConfig{}, fmt.Errorf(
			`config: %s %q must be a host:port address such as ":8080": %w`,
			httpListenAddressEnvVar,
			listenAddress,
			err,
		)
	}
	if strings.TrimSpace(host) != host {
		return HTTPConfig{}, fmt.Errorf(`config: %s %q must not contain surrounding whitespace`, httpListenAddressEnvVar, listenAddress)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return HTTPConfig{}, fmt.Errorf(`config: %s %q must use a port between 1 and 65535`, httpListenAddressEnvVar, listenAddress)
	}
	return HTTPConfig{ListenAddress: listenAddress}, nil
}

func loadProviderKind() (provider.Kind, error) {
	providerName, isConfigured := os.LookupEnv(providerEnvVar)
	if !isConfigured || providerName == "" {
		return provider.KindNVML, nil
	}
	switch provider.Kind(providerName) {
	case provider.KindNVML, provider.KindReplay:
		return provider.Kind(providerName), nil
	default:
		return "", fmt.Errorf("config: %s must be nvml or replay", providerEnvVar)
	}
}
