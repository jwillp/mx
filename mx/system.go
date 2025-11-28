package mx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/morebec/misas/misas"
)

type SystemInfo struct {
	Name        string
	Version     string
	Environment Environment
	Debug       bool
}

type System struct {
	info           SystemInfo
	logger         *slog.Logger
	clock          misas.Clock
	pm             PluginManager
	customPlugins  []Plugin
	builtInPlugins []Plugin
}

func newSystem(sc SystemConf) *System {

	return &System{
		info: SystemInfo{
			Name:        sc.name,
			Version:     sc.version,
			Environment: sc.environment,
			Debug:       sc.debug,
		},
		clock:          sc.clock,
		logger:         slog.New(sc.loggerHandler),
		pm:             newPluginManager(),
		builtInPlugins: []Plugin{loggingPlugin{}},
		customPlugins:  sc.plugins,
	}
}

func (s *System) run(app ApplicationSubsystem) error {
	if err := s.doRun(app); err != nil {
		return fmt.Errorf("system failed: %w", err)
	}

	return nil
}

func (s *System) doRun(app ApplicationSubsystem) error {
	ctx := newSystemContext(*s)

	ctx, cancel := s.setupSignalHandling(ctx)
	defer cancel()

	s.loadPlugins(ctx, app)

	// Wrap app with management layer
	app = newManagedApplicationSubsystem(app, s.pm, s.clock)
	appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})

	// Setup teardown with a fresh context (not the canceled one)
	defer s.teardownApplication(newSystemContext(*s), appCtx, app)

	if err := s.initializeApplication(ctx, appCtx, app); err != nil {
		return fmt.Errorf("failed to initialize application %q: %w", app.Name(), err)
	}

	if err := s.runApplication(ctx, appCtx, app); err != nil {
		return fmt.Errorf("application %q failed during execution: %w", app.Name(), err)
	}

	return nil
}

func (s *System) setupSignalHandling(ctx context.Context) (context.Context, context.CancelFunc) {
	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)

	// Monitor for termination signals in a separate goroutine
	go func() {
		defer signal.Stop(signalChan)
		sig := <-signalChan
		Log(ctx).Info("received termination signal, initiating graceful shutdown", "signal", sig.String())
		cancel()
	}()

	return ctx, cancel
}

func (s *System) loadPlugins(ctx context.Context, app ApplicationSubsystem) {
	// Collect all plugins: built-in, custom, and app-provided
	allPlugins := make([]Plugin, 0, len(s.builtInPlugins)+len(s.customPlugins)+1)
	allPlugins = append(allPlugins, s.builtInPlugins...)
	allPlugins = append(allPlugins, s.customPlugins...)
	if plugin, ok := app.(Plugin); ok {
		allPlugins = append(allPlugins, plugin)
	}

	// Add all plugins to the manager
	for _, plugin := range allPlugins {
		s.pm.AddPlugin(ctx, plugin)
	}
}

func (s *System) initializeApplication(ctx context.Context, appCtx context.Context, app ApplicationSubsystem) error {
	// Dispatch initialization started hook
	initializationStartedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, SystemInitializationStartedHook{
		Name:        s.info.Name,
		Version:     s.info.Version,
		Environment: s.info.Environment,
		Debug:       s.info.Debug,
		StartedAt:   initializationStartedAt,
		System:      s,
	})

	// Initialize the application subsystem
	err := app.Initialize(appCtx)
	s.pm.DispatchHook(ctx, SystemInitializationEndedHook{
		StartedAt: initializationStartedAt,
		EndedAt:   s.clock.Now(),
		Error:     err,
	})

	return err
}

func (s *System) runApplication(ctx context.Context, appCtx context.Context, app ApplicationSubsystem) error {
	// Dispatch run started hook
	runStartedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, SystemRunStartedHook{StartedAt: runStartedAt})

	// Run the application subsystem
	err := app.Run(appCtx)
	s.pm.DispatchHook(ctx, SystemRunEndedHook{
		StartedAt: runStartedAt,
		EndedAt:   s.clock.Now(),
		Error:     err,
	})

	return err
}

func (s *System) teardownApplication(ctx context.Context, appCtx context.Context, app ApplicationSubsystem) {
	// Dispatch teardown started hook
	teardownStartedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, SystemTeardownStartedHook{StartedAt: teardownStartedAt})

	// Create a fresh context for teardown (not the canceled one)
	teardownCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})

	// Teardown the application subsystem
	teardownErr := app.Teardown(teardownCtx)
	s.pm.DispatchHook(ctx, SystemTeardownEndedHook{
		StartedAt: teardownStartedAt,
		EndedAt:   s.clock.Now(),
		Error:     teardownErr,
	})
}

func (s *System) PluginManager() PluginManager { return s.pm }
func (s *System) Clock() misas.Clock           { return s.clock }

type SubsystemInfo struct {
	Name string
}
