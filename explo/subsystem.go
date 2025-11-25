package explo

type BusinessSubsystem struct {
	Name            string
	commandHandlers map[CommandTypeName]any
}

func NewBusinessSubsystem(name string) *BusinessSubsystem {
	return &BusinessSubsystem{Name: name}
}

func (s *BusinessSubsystem) WithCommandHandler(ct CommandTypeName, handler any) *BusinessSubsystem {
	if s.commandHandlers == nil {
		s.commandHandlers = make(map[CommandTypeName]any)
	}

	s.commandHandlers[ct] = handler

	return s
}

func (s *BusinessSubsystem) CommandHandlers() map[CommandTypeName]any { return s.commandHandlers }
