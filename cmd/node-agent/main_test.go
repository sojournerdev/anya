package main

import (
	"context"
	"testing"
)

func TestRunBuildsReplayApplicationAndStopsWhenCanceled(t *testing.T) {
	t.Setenv("AGENT_PROVIDER", "replay")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := run(ctx); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}
func TestRunRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("AGENT_REFRESH_INTERVAL", "fast")

	if err := run(context.Background()); err == nil {
		t.Fatal("run() succeeded with invalid configuration")
	}
}
func TestRunRejectsInvalidConfigurationBeforeProviderStartup(t *testing.T) {
	t.Setenv("AGENT_REFRESH_INTERVAL", "1m")
	t.Setenv("AGENT_MAX_STATE_AGE", "30s")

	if err := run(context.Background()); err == nil {
		t.Fatal("run() succeeded with an invalid freshness configuration")
	}
}
