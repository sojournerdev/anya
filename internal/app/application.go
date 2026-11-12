// Package app runs the node-agent application lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sojournerdev/anya/internal/agent"
)

const applicationShutdownTimeout = 5 * time.Second

// Dependencies contains the constructed components used by the application.
type Dependencies struct {
	Agent      *agent.Agent
	HTTPServer *http.Server
}

// Settings contains validated application runtime settings.
type Settings struct {
	RefreshInterval time.Duration
	RefreshTimeout  time.Duration
	ListenPort      int
}

// Application owns refresh scheduling and process lifecycle.
type Application struct {
	nodeAgent       *agent.Agent
	refreshInterval time.Duration
	refreshTimeout  time.Duration
	httpServer      *http.Server
	listenPort      int
}

// New creates an application from constructed dependencies and validated settings.
func New(dependencies Dependencies, settings Settings) (*Application, error) {
	if dependencies.Agent == nil {
		return nil, errors.New("app: agent is required")
	}
	if dependencies.HTTPServer == nil {
		return nil, errors.New("app: HTTP server is required")
	}
	if settings.RefreshInterval <= 0 {
		return nil, errors.New("app: refresh interval must be positive")
	}
	if settings.RefreshTimeout <= 0 {
		return nil, errors.New("app: refresh timeout must be positive")
	}
	if settings.RefreshTimeout > settings.RefreshInterval {
		return nil, errors.New("app: refresh timeout must not exceed refresh interval")
	}
	if settings.ListenPort < 0 || settings.ListenPort > 65535 {
		return nil, errors.New("app: HTTP port must be between 0 and 65535")
	}
	return &Application{
		nodeAgent:       dependencies.Agent,
		refreshInterval: settings.RefreshInterval,
		refreshTimeout:  settings.RefreshTimeout,
		httpServer:      dependencies.HTTPServer,
		listenPort:      settings.ListenPort,
	}, nil
}

// Run starts the application until cancellation or a component fails.
func (a *Application) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("app: context is required")
	}
	if ctx.Err() != nil {
		// Cancellation before startup is a normal shutdown, not a failure.
		return nil //nolint:nilerr
	}

	listener, err := a.openHTTPListener()
	if err != nil {
		return err
	}

	componentError := a.runComponents(ctx, listener)
	return a.shutdown(ctx, componentError)
}

func (a *Application) openHTTPListener() (net.Listener, error) {
	address := net.JoinHostPort("", strconv.Itoa(a.listenPort))
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		return nil, fmt.Errorf("listen for HTTP server on %q: %w", address, err)
	}
	log.Printf("HTTP server listening on %s", listener.Addr())
	return listener, nil
}

func (a *Application) runComponents(ctx context.Context, listener net.Listener) error {
	runContext, cancel := context.WithCancel(ctx)
	defer cancel()

	httpServerErrors := make(chan error, 1)
	go func() {
		httpServerErrors <- a.httpServer.Serve(listener)
	}()

	refreshLoopErrors := make(chan error, 1)
	go func() {
		refreshLoopErrors <- a.runRefreshLoop(runContext)
	}()

	var componentError error
	select {
	case componentError = <-httpServerErrors:
		if !errors.Is(componentError, http.ErrServerClosed) {
			componentError = fmt.Errorf("serve HTTP: %w", componentError)
		} else {
			componentError = nil
		}
	case componentError = <-refreshLoopErrors:
		if componentError != nil {
			componentError = fmt.Errorf("run agent refresh loop: %w", componentError)
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	return componentError
}

func (a *Application) shutdown(ctx context.Context, componentError error) error {
	shutdownContext, cancelShutdown := context.WithTimeout(
		context.WithoutCancel(ctx),
		applicationShutdownTimeout,
	)
	defer cancelShutdown()
	if shutdownError := a.httpServer.Shutdown(shutdownContext); shutdownError != nil && componentError == nil {
		componentError = fmt.Errorf("shut down HTTP server: %w", shutdownError)
	}
	if closeError := a.nodeAgent.Close(); closeError != nil && componentError == nil {
		componentError = fmt.Errorf("close agent: %w", closeError)
	}
	return componentError
}

func (a *Application) runRefreshLoop(ctx context.Context) error {
	refresh := func() {
		refreshContext, cancel := context.WithTimeout(ctx, a.refreshTimeout)
		defer cancel()
		if err := a.nodeAgent.Refresh(refreshContext); err != nil && ctx.Err() == nil {
			log.Printf("GPU refresh failed: %v", err)
		}
	}

	refresh()
	ticker := time.NewTicker(a.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			refresh()
		}
	}
}
