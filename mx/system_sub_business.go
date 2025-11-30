package mx

import (
	"context"

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
		panic("business subsystem: " + bc.name + " command type name cannot be empty")
	}
	if h == nil {
		panic("business subsystem: " + bc.name + " command handler cannot be nil")
	}
	bc.commandHandlers[ct] = h
	return bc
}

func (bc *BusinessSubsystemConf) WithEventHandlers(eventBusName EventBusName, handlers ...misas.EventHandler) *BusinessSubsystemConf {
	if eventBusName == "" {
		panic("business subsystem: " + bc.name + " event bus name cannot be empty")
	}

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
