package mx

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/morebec/misas/misas"
)

// withCommandContextPropagation wraps a command handler to propagate subsystem context.
func withCommandContextPropagation(subsystemName string, h misas.CommandHandler) misas.CommandHandler {
	return misas.CommandHandlerFunc(func(ctx context.Context, cmd misas.Command) misas.CommandResult {
		ctx = newSubsystemContext(ctx, SubsystemInfo{Name: subsystemName})
		return h.Handle(ctx, cmd)
	})
}

// withCommandLogging wraps a command handler to log command execution.
func withCommandLogging(h misas.CommandHandler) misas.CommandHandler {
	return misas.CommandHandlerFunc(func(ctx context.Context, cmd misas.Command) misas.CommandResult {
		logger := Log(ctx)

		origin := Ctx(ctx).SubsystemOrigin()
		if (origin != SubsystemInfo{}) {
			logger = logger.With(slog.String("originSubsystem", origin.Name))
		}

		logger.Info(fmt.Sprintf("handling command %q", cmd.TypeName()), slog.String("command", string(cmd.TypeName())))
		result := h.Handle(ctx, cmd)
		if result.Error != nil {
			logger.Error(
				fmt.Sprintf("failed to handle command %q", cmd.TypeName()),
				slog.Any(logKeyError, result.Error),
				slog.String("command", string(cmd.TypeName())),
			)
		} else {
			logger.Info(fmt.Sprintf("command %q handled successfully", cmd.TypeName()), slog.String("command", string(cmd.TypeName())))
		}

		return result
	})
}

// withQueryContextPropagation wraps a query handler to propagate subsystem context.
func withQueryContextPropagation(subsystemName string, h misas.QueryHandler) misas.QueryHandler {
	return misas.QueryHandlerFunc(func(ctx context.Context, q misas.Query) misas.QueryResult {
		ctx = newSubsystemContext(ctx, SubsystemInfo{Name: subsystemName})
		return h.Handle(ctx, q)
	})
}

// withQueryLogging wraps a query handler to log query execution.
func withQueryLogging(h misas.QueryHandler) misas.QueryHandler {
	return misas.QueryHandlerFunc(func(ctx context.Context, q misas.Query) misas.QueryResult {
		logger := Log(ctx)

		origin := Ctx(ctx).SubsystemOrigin()
		if (origin != SubsystemInfo{}) {
			logger = logger.With(slog.String("originSubsystem", origin.Name))
		}

		logger.Info(fmt.Sprintf("handling query %q", q.TypeName()), slog.String("query", string(q.TypeName())))
		result := h.Handle(ctx, q)

		if result.Error != nil {
			logger.Error(
				fmt.Sprintf("failed to handle query %q", q.TypeName()),
				slog.Any(logKeyError, result.Error),
				slog.String("query", string(q.TypeName())),
			)
		} else {
			logger.Info(fmt.Sprintf("successfully handled query %q", q.TypeName()), slog.String("query", string(q.TypeName())))
		}

		return result
	})
}

// withEventContextPropagation wraps an event handler to propagate subsystem context.
func withEventContextPropagation(subsystemName string, h misas.EventHandler) misas.EventHandler {
	return misas.EventHandlerFunc(func(ctx context.Context, e misas.Event) error {
		ctx = newSubsystemContext(ctx, SubsystemInfo{Name: subsystemName})
		return h.Handle(ctx, e)
	})
}

// withEventLogging wraps an event handler to log event handling.
func withEventLogging(h misas.EventHandler) misas.EventHandler {
	return misas.EventHandlerFunc(func(ctx context.Context, e misas.Event) error {
		logger := Log(ctx)

		origin := Ctx(ctx).SubsystemOrigin()
		if (origin != SubsystemInfo{}) {
			logger = logger.With(slog.String("originSubsystem", origin.Name))
		}

		logger.Info(fmt.Sprintf("handling event %q", e.TypeName()))
		err := h.Handle(ctx, e)
		if err != nil {
			logger.Error(
				fmt.Sprintf("failed to handle event %q", e.TypeName()),
				slog.Any(logKeyError, err),
				slog.String("event", string(e.TypeName())),
			)
		} else {
			logger.Info(fmt.Sprintf("event %q handled successfully", e.TypeName()))
		}
		return err
	})
}
