package explo

import "context"

type CommandTypeName string

type Command interface{ TypeName() CommandTypeName }

type CommandHandler[C Command] interface {
	Handle(context.Context, C) error
}
