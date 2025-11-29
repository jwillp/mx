package misas

import (
	"context"
	"sync"
)

type CommandTypeName string
type Command interface{ TypeName() CommandTypeName }
type CommandResult struct {
	Payload any
}
type CommandHandler interface {
	Handle(context.Context, Command) CommandResult
}

type CommandHandlerFunc func(context.Context, Command) CommandResult

func (f CommandHandlerFunc) Handle(ctx context.Context, cmd Command) CommandResult {
	return f(ctx, cmd)
}

type TypedCommandHandler[T Command] interface {
	Handle(context.Context, T) CommandResult
}

type CommandBus interface {
	HandleCommand(context.Context, Command) CommandResult
	RegisterHandler(CommandTypeName, CommandHandler)
}

type InMemoryCommandBus struct {
	handlers map[CommandTypeName]CommandHandler
	mu       sync.Mutex
}

func NewInMemoryCommandBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{
		handlers: make(map[CommandTypeName]CommandHandler),
	}
}

func (b *InMemoryCommandBus) HandleCommand(ctx context.Context, cmd Command) CommandResult {
	if cmd == nil {
		panic("command cannot be nil") // TODO error
		//return CommandResult{
		//	Error: ErrNilCommand{},
		//}
	}
	b.mu.Lock()
	handler, ok := b.handlers[cmd.TypeName()]
	b.mu.Unlock()
	if !ok {
		panic("no command handler registered for command type: " + string(cmd.TypeName())) // TODO error
		//return CommandResult{
		//	Error: ErrNoCommandHandlerRegistered{CommandTypeName: cmd.TypeName()},
		//}
	}

	return handler.Handle(ctx, cmd)
}

func (b *InMemoryCommandBus) RegisterHandler(cmdType CommandTypeName, handler CommandHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[cmdType] = handler
}
