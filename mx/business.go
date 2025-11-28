package mx

import (
	"context"
	"github.com/morebec/misas/misas"
)

type BusinessSubsystemConf struct {
	name            string
	commandHandlers map[misas.CommandTypeName]misas.CommandHandler
	eventHandlers   map[string][]misas.EventHandler
}

func NewBusinessSubsystem(name string) *BusinessSubsystemConf {
	return &BusinessSubsystemConf{
		name:            name,
		commandHandlers: make(map[misas.CommandTypeName]misas.CommandHandler),
		eventHandlers:   make(map[string][]misas.EventHandler),
	}
}

func (bc *BusinessSubsystemConf) WithCommandHandler(ct misas.CommandTypeName, h misas.CommandHandler) *BusinessSubsystemConf {
	bc.commandHandlers[ct] = h
	return bc
}

func (bc *BusinessSubsystemConf) WithEventHandlers(eventBusName string, handlers ...misas.EventHandler) *BusinessSubsystemConf {
	bc.eventHandlers[eventBusName] = append(bc.eventHandlers[eventBusName], handlers...)
	return bc
}

type InMemoryCommandBus struct {
	handlers map[misas.CommandTypeName]misas.CommandHandler
}

func NewInMemoryCommandBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{
		handlers: make(map[misas.CommandTypeName]misas.CommandHandler),
	}
}

func (b *InMemoryCommandBus) HandleCommand(ctx context.Context, cmd misas.Command) misas.CommandResult {
	if cmd == nil {
		panic("command cannot be nil") // TODO error
		//return misas.CommandResult{
		//	Error: misas.ErrNilCommand{},
		//}
	}
	handler, ok := b.handlers[cmd.TypeName()]
	if !ok {
		panic("no command handler registered for command type: " + string(cmd.TypeName())) // TODO error
		//return misas.CommandResult{
		//	Error: misas.ErrNoCommandHandlerRegistered{CommandTypeName: cmd.TypeName()},
		//}
	}

	return handler.Handle(ctx, cmd)
}

func (b *InMemoryCommandBus) RegisterHandler(cmdType misas.CommandTypeName, handler misas.CommandHandler) {
	b.handlers[cmdType] = handler
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
