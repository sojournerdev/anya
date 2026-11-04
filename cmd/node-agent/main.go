// Package main composes and runs the Anya node agent.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sojournerdev/anya/internal/agent"
	"github.com/sojournerdev/anya/internal/app"
	"github.com/sojournerdev/anya/internal/config"
	"github.com/sojournerdev/anya/internal/provider/factory"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	observationProvider, err := factory.New(settings.Provider)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}
	nodeAgent, err := agent.New(observationProvider)
	if err != nil {
		_ = observationProvider.Close()
		return fmt.Errorf("create agent: %w", err)
	}

	httpServer, err := app.NewServer(nodeAgent, settings.Refresh.MaxAge)
	if err != nil {
		_ = nodeAgent.Close()
		return fmt.Errorf("create HTTP server: %w", err)
	}
	agentApplication, err := app.New(
		app.Dependencies{
			Agent:      nodeAgent,
			HTTPServer: httpServer},
		app.Settings{
			RefreshInterval: settings.Refresh.Interval,
			RefreshTimeout:  settings.Refresh.Timeout,
			ListenAddress:   settings.HTTP.ListenAddress},
	)
	if err != nil {
		_ = nodeAgent.Close()
		return fmt.Errorf("create application: %w", err)
	}
	return agentApplication.Run(ctx)
}
