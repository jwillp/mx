package misas

import "context"

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
