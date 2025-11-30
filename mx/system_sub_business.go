package mx

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"log/slog"

	"github.com/morebec/misas/misas"
)

type EventBusName string

type BusinessSubsystemConf struct {
	name            string
	commandHandlers map[misas.CommandTypeName]misas.CommandHandler
	eventHandlers   map[EventBusName][]misas.EventHandler
}

func NewBusinessSubsystem(name string) *BusinessSubsystemConf {
	if name == "" {
		panic("business subsystem name cannot be empty")
	}
	return &BusinessSubsystemConf{
		name:            name,
		commandHandlers: make(map[misas.CommandTypeName]misas.CommandHandler),
		eventHandlers:   make(map[EventBusName][]misas.EventHandler),
	}
}

func (bc *BusinessSubsystemConf) WithCommandHandler(ct misas.CommandTypeName, h misas.CommandHandler) *BusinessSubsystemConf {
	if ct == "" {
		panic(fmt.Sprintf("business subsystem %s: command type name cannot be empty", bc.name))
	}
	if h == nil {
		panic(fmt.Sprintf("business subsystem %s: handler cannot be nil", bc.name))
	}
	bc.commandHandlers[ct] = systemCommandHandler(bc.name, h)
	return bc
}

func (bc *BusinessSubsystemConf) WithEventHandlers(eventBusName EventBusName, handlers ...misas.EventHandler) *BusinessSubsystemConf {
	if eventBusName == "" {
		panic(fmt.Sprintf("business subsystem %s: event bus name cannot be empty", bc.name))
	}

	handlers = lo.Map(handlers, func(h misas.EventHandler, _ int) misas.EventHandler {
		return newSystemEventHandler(bc.name, h)
	})
	bc.eventHandlers[eventBusName] = append(bc.eventHandlers[eventBusName], handlers...)

	return bc
}

type DynamicBindingCommandBus struct {
	*DynamicBinding[misas.CommandBus]
}

func NewDynamicBindingCommandBus() *DynamicBindingCommandBus {
	return &DynamicBindingCommandBus{DynamicBinding: NewDynamicBinding[misas.CommandBus]()}
}

func (d *DynamicBindingCommandBus) HandleCommand(ctx context.Context, cmd misas.Command) misas.CommandResult {
	return d.Get().HandleCommand(ctx, cmd)
}

func (d *DynamicBindingCommandBus) RegisterHandler(cmdType misas.CommandTypeName, handler misas.CommandHandler) {
	d.Get().RegisterHandler(cmdType, handler)
}

type DynamicBindingEventBus struct {
	*DynamicBinding[misas.EventBus]
}

func NewDynamicBindingEventBus() *DynamicBindingEventBus {
	return &DynamicBindingEventBus{DynamicBinding: NewDynamicBinding[misas.EventBus]()}
}

func (d DynamicBindingEventBus) RegisterHandler(handler misas.EventHandler) {
	d.Get().RegisterHandler(handler)
}

func (d DynamicBindingEventBus) Publish(ctx context.Context, event misas.Event) error {
	return d.Get().Publish(ctx, event)
}

func systemCommandHandler(subsystemName string, h misas.CommandHandler) misas.CommandHandler {
	return misas.CommandHandlerFunc(func(ctx context.Context, cmd misas.Command) misas.CommandResult {
		origin := Ctx(ctx).SubsystemInfo().Name
		ctx = newSubsystemContext(ctx, SubsystemInfo{Name: subsystemName})
		logger := Log(ctx).With(slog.String("originSubsystem", origin))

		logger.Info(fmt.Sprintf("handling command %q", cmd.TypeName()), slog.String("command", string(cmd.TypeName())))
		result := h.Handle(ctx, cmd)
		if result.Error != nil {
			logger.Error(
				fmt.Sprintf("failed to handle command %q", cmd.TypeName()),
				slog.Any(logKeyError, result.Error),
				slog.String(
					"command",
					string(cmd.TypeName()),
				),
			)
		} else {
			logger.Info(fmt.Sprintf("command %q handled successfully", cmd.TypeName()), slog.String("command", string(cmd.TypeName())))
		}

		return result
	})
}

func newSystemEventHandler(subsystemName string, h misas.EventHandler) misas.EventHandler {
	return misas.EventHandlerFunc(func(ctx context.Context, e misas.Event) error {
		origin := Ctx(ctx).SubsystemInfo().Name
		ctx = newSubsystemContext(ctx, SubsystemInfo{Name: subsystemName})
		logger := Log(ctx).With(slog.String("originSubsystem", origin))

		logger.Info(
			fmt.Sprintf("handling event %q", e.TypeName()),
		)
		err := h.Handle(ctx, e)
		if err != nil {
			logger.Error(
				fmt.Sprintf("failed to handle command %q", e.TypeName()),
				slog.Any(logKeyError, err),
				slog.String(
					"command",
					string(e.TypeName()),
				),
			)
		} else {
			logger.Info(fmt.Sprintf("event %q handled successfully", e.TypeName()))
		}
		return err
	})
}
