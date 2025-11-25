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

type TypedCommandHandler[T Command] interface {
	Handle(context.Context, T) CommandResult
}
