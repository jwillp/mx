package mx

import (
	"context"
	"fmt"
	"github.com/morebec/misas/misas"
	"github.com/samber/lo"
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

// WithCommandHandler registers a command handler for the given command type with the system's command bus.
// It also automatically the command type in the global command registry for serialization purposes.
func (bc *BusinessSubsystemConf) WithCommandHandler(ct misas.Command, h misas.CommandHandler) *BusinessSubsystemConf {
	if ct == nil {
		panic(fmt.Sprintf("business subsystem %s: command cannot be empty", bc.name))
	}
	if h == nil {
		panic(fmt.Sprintf("business subsystem %s: handler cannot be nil", bc.name))
	}
	h = withCommandLogging(h)
	h = withCommandContextPropagation(bc.name, h)
	bc.commandHandlers[ct.TypeName()] = h

	CommandRegistry.Register(ct.TypeName(), ct)

	return bc
}

// WithEventHandlers registers event handlers for the given event bus name with the system's event buses.
func (bc *BusinessSubsystemConf) WithEventHandlers(eventBusName EventBusName, handlers ...misas.EventHandler) *BusinessSubsystemConf {
	if eventBusName == "" {
		panic(fmt.Sprintf("business subsystem %s: event bus name cannot be empty", bc.name))
	}

	handlers = lo.Map(handlers, func(h misas.EventHandler, _ int) misas.EventHandler {
		h = withEventLogging(h)
		h = withEventContextPropagation(bc.name, h)
		return h
	})
	bc.eventHandlers[eventBusName] = append(bc.eventHandlers[eventBusName], handlers...)

	return bc
}

// ProducesEvents registers the given events in the global event registry for serialization purposes.
func (bc *BusinessSubsystemConf) ProducesEvents(events ...misas.Event) *BusinessSubsystemConf {
	for _, e := range events {
		EventRegistry.Register(e.TypeName(), e)
	}

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
