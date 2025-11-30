package mx

import (
	"context"
	"fmt"
	"github.com/morebec/misas/mtime"
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
	info               SystemInfo
	logger             *slog.Logger
	clock              mtime.Clock
	pm                 SystemPluginManager
	builtInPlugins     []SystemPlugin
	customPlugins      []SystemPlugin
	commandBus         misas.CommandBus
	eventBuses         map[EventBusName]misas.EventBus
	businessSubsystems map[string]BusinessSubsystemConf
	queryBus           misas.QueryBus
	querySubsystems    map[string]QuerySubsystemConf
}

func newSystem(sc *SystemConf) *System {
	if !sc.commandBus.IsBound() {
		sc.commandBus.Bind(misas.NewInMemoryCommandBus())
	}

	for _, eventBus := range sc.eventBuses {
		if !eventBus.IsBound() {
			eventBus.Bind(misas.NewInMemoryEventBus())
		}
	}

	if !sc.queryBus.IsBound() {
		sc.queryBus.Bind(misas.NewInMemoryQueryBus())
	}

	// Collect event buses for the system
	eventBuses := make(map[EventBusName]misas.EventBus, len(sc.eventBuses))
	for name, eb := range sc.eventBuses {
		eventBuses[name] = eb
	}

	return &System{
		info: SystemInfo{
			Name:        sc.name,
			Version:     sc.version,
			Environment: sc.environment,
			Debug:       sc.debug,
		},
		clock:              sc.clock,
		logger:             slog.New(sc.loggerHandler),
		pm:                 newPluginManager(),
		builtInPlugins:     []SystemPlugin{loggingPlugin{}},
		customPlugins:      sc.plugins,
		commandBus:         sc.commandBus,
		eventBuses:         eventBuses,
		businessSubsystems: sc.businessSubsystems,
		queryBus:           sc.queryBus,
		querySubsystems:    sc.querySubsystems,
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
	defer s.teardownSystem(newSystemContext(*s), app)

	if err := s.initializeSystem(ctx, appCtx, app); err != nil {
		return fmt.Errorf("failed to initialize application %q: %w", app.Name(), err)
	}

	if err := s.executeSystem(ctx, appCtx, app); err != nil {
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
	allPlugins := make([]SystemPlugin, 0, len(s.builtInPlugins)+len(s.customPlugins)+1)
	allPlugins = append(allPlugins, s.builtInPlugins...)
	allPlugins = append(allPlugins, s.customPlugins...)
	if plugin, ok := app.(SystemPlugin); ok {
		allPlugins = append(allPlugins, plugin)
	}

	// Add all plugins to the manager
	for _, plugin := range allPlugins {
		s.pm.AddPlugin(ctx, plugin)
	}
}

func (s *System) initializeBusinessSubsystems(ctx context.Context) {
	for _, bsConf := range s.businessSubsystems {
		bsCtx := newSubsystemContext(ctx, SubsystemInfo{Name: bsConf.name})

		// Dispatch business subsystem initialization started hook
		initStartedAt := s.clock.Now()
		s.pm.DispatchHook(bsCtx, BusinessSubsystemInitializationStartedHook{
			BusinessSubsystemName: bsConf.name,
			StartedAt:             initStartedAt,
		})

		// Register command handlers
		for cmdType, handler := range bsConf.commandHandlers {
			s.commandBus.RegisterHandler(cmdType, handler)
		}

		// Register event handlers
		s.registerEventHandlers(bsCtx, bsConf.eventHandlers)

		// Dispatch business subsystem initialization ended hook
		s.pm.DispatchHook(bsCtx, BusinessSubsystemInitializationEndedHook{
			BusinessSubsystemName: bsConf.name,
			StartedAt:             initStartedAt,
			EndedAt:               s.clock.Now(),
			Error:                 nil,
		})
	}
}

func (s *System) initializeQuerySubsystems(ctx context.Context) {
	for _, qsConf := range s.querySubsystems {
		qsCtx := newSubsystemContext(ctx, SubsystemInfo{Name: qsConf.name})

		// Dispatch query subsystem initialization started hook
		initStartedAt := s.clock.Now()
		s.pm.DispatchHook(qsCtx, QuerySubsystemInitializationStartedHook{
			QuerySubsystemName: qsConf.name,
			StartedAt:          initStartedAt,
		})

		// Register query handlers
		for queryType, handler := range qsConf.queryHandlers {
			s.queryBus.RegisterHandler(queryType, handler)
		}

		// Register event handlers
		s.registerEventHandlers(qsCtx, qsConf.eventHandlers)

		// Dispatch query subsystem initialization ended hook
		s.pm.DispatchHook(qsCtx, QuerySubsystemInitializationEndedHook{
			QuerySubsystemName: qsConf.name,
			StartedAt:          initStartedAt,
			EndedAt:            s.clock.Now(),
			Error:              nil,
		})
	}
}

func (s *System) registerEventHandlers(qsCtx context.Context, handlers map[EventBusName][]misas.EventHandler) {
	for eventBusName, busHandlers := range handlers {
		eb, ok := s.eventBuses[eventBusName]
		if !ok {
			Log(qsCtx).Warn(fmt.Sprintf(
				"some event handler(s) are subscribed to event bus %q, but it does not publish events; skipping registration...",
				eventBusName,
			)) // This message can be suppressed by ensuring a call to system.EventBus(eventBusName)
			continue
		}
		for _, h := range busHandlers {
			eb.RegisterHandler(h)
		}
	}
}

func (s *System) initializeSystem(ctx context.Context, appCtx context.Context, app ApplicationSubsystem) error {
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

	s.initializeBusinessSubsystems(ctx)

	s.initializeQuerySubsystems(ctx)

	// Initialize the application subsystem
	err := app.Initialize(appCtx)
	s.pm.DispatchHook(ctx, SystemInitializationEndedHook{
		StartedAt: initializationStartedAt,
		EndedAt:   s.clock.Now(),
		Error:     err,
	})

	return err
}

func (s *System) executeSystem(ctx context.Context, appCtx context.Context, app ApplicationSubsystem) error {
	// Dispatch run started hook
	runStartedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, SystemExecutionStartedHook{StartedAt: runStartedAt})

	// Run the application subsystem
	err := app.Run(appCtx)
	s.pm.DispatchHook(ctx, SystemExecutionEndedHook{
		StartedAt: runStartedAt,
		EndedAt:   s.clock.Now(),
		Error:     err,
	})

	return err
}

func (s *System) teardownSystem(ctx context.Context, app ApplicationSubsystem) {
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

func (s *System) PluginManager() SystemPluginManager { return s.pm }
func (s *System) Clock() mtime.Clock                 { return s.clock }

type SubsystemInfo struct {
	Name string
}
