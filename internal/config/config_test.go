package config_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sojournerdev/anya/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		environment   map[string]string
		wantConfig    config.Config
		wantErrorText string
	}{
		{
			name:        "defaults",
			environment: map[string]string{},
			wantConfig: config.Config{
				Refresh: config.RefreshConfig{
					Interval: config.DefaultRefreshInterval,
					Timeout:  config.DefaultRefreshTimeout,
					MaxAge:   config.DefaultMaxStateAge,
				},
				HTTP:     config.HTTPConfig{Port: config.DefaultHTTPPort},
				Provider: "nvml",
			},
		},
		{
			name: "configured settings",
			environment: map[string]string{
				"AGENT_REFRESH_INTERVAL": "2s",
				"AGENT_REFRESH_TIMEOUT":  "1s",
				"AGENT_MAX_STATE_AGE":    "20s",
				"AGENT_HTTP_PORT":        "9090",
			},
			wantConfig: config.Config{
				Refresh: config.RefreshConfig{
					Interval: 2 * time.Second,
					Timeout:  time.Second,
					MaxAge:   20 * time.Second,
				},
				HTTP:     config.HTTPConfig{Port: 9090},
				Provider: "nvml",
			},
		},
		{
			name: "invalid refresh interval",
			environment: map[string]string{
				"AGENT_REFRESH_INTERVAL": "fast",
			},
			wantErrorText: "AGENT_REFRESH_INTERVAL",
		},
		{
			name: "refresh interval below minimum",
			environment: map[string]string{
				"AGENT_REFRESH_INTERVAL": "500ms",
			},
			wantErrorText: "outside the allowed range",
		},
		{
			name: "refresh timeout exceeds refresh interval",
			environment: map[string]string{
				"AGENT_REFRESH_TIMEOUT":  "6s",
				"AGENT_REFRESH_INTERVAL": "5s",
			},
			wantErrorText: "must not exceed interval",
		},
		{
			name: "default max state age shorter than configured refresh interval",
			environment: map[string]string{
				"AGENT_REFRESH_INTERVAL": "1m",
			},
			wantErrorText: "must be at least interval",
		},
		{
			name: "max state age shorter than refresh interval",
			environment: map[string]string{
				"AGENT_MAX_STATE_AGE": "1s",
			},
			wantErrorText: "must be at least interval",
		},
		{
			name: "max state age above maximum",
			environment: map[string]string{
				"AGENT_MAX_STATE_AGE": "2h",
			},
			wantErrorText: "outside the allowed range",
		},
		{
			name: "invalid max state age",
			environment: map[string]string{
				"AGENT_MAX_STATE_AGE": "later",
			},
			wantErrorText: "AGENT_MAX_STATE_AGE",
		},
		{
			name: "unknown provider",
			environment: map[string]string{
				"AGENT_PROVIDER": "unknown",
			},
			wantErrorText: "AGENT_PROVIDER",
		},
		{
			name: "invalid HTTP port",
			environment: map[string]string{
				"AGENT_HTTP_PORT": "invalid",
			},
			wantErrorText: "AGENT_HTTP_PORT",
		},
		{
			name: "HTTP port outside range",
			environment: map[string]string{
				"AGENT_HTTP_PORT": "70000",
			},
			wantErrorText: "between 1 and 65535",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setEnvironment(t, test.environment)

			got, err := config.Load()
			if test.wantErrorText != "" {
				if err == nil {
					t.Fatal("Load() succeeded, want error")
				}
				if !strings.Contains(err.Error(), test.wantErrorText) {
					t.Fatalf("Load() error = %q, want context %q", err, test.wantErrorText)
				}
				if got != (config.Config{}) {
					t.Fatalf("Load() returned partial config: %#v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got != test.wantConfig {
				t.Fatalf("Load() = %#v, want %#v", got, test.wantConfig)
			}
		})
	}
}

func setEnvironment(t *testing.T, values map[string]string) {
	t.Helper()
	for _, name := range []string{
		"AGENT_REFRESH_INTERVAL",
		"AGENT_REFRESH_TIMEOUT",
		"AGENT_MAX_STATE_AGE",
		"AGENT_HTTP_PORT",
		"AGENT_PROVIDER",
	} {
		if value, configured := values[name]; configured {
			t.Setenv(name, value)
			continue
		}
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("Unsetenv(%q) error = %v", name, err)
		}
	}
}
