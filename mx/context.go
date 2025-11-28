package mx

import (
	"context"
)

type systemLoggerContextKey struct{}

type systemInfoContextKey struct{}
type subsystemInfoContextKey struct{}

func newSystemContext(s System) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, systemLoggerContextKey{}, s.logger)
	ctx = context.WithValue(ctx, subsystemInfoContextKey{}, s.info)

	return ctx
}

func newSubsystemContext(ctx context.Context, info SubsystemInfo) context.Context {
	ctx = context.WithValue(ctx, subsystemInfoContextKey{}, info)

	return ctx
}

func GetSubsystemInfoFromContext(ctx context.Context) SubsystemInfo {
	subsystem, ok := ctx.Value(subsystemInfoContextKey{}).(SubsystemInfo)
	if !ok {
		return SubsystemInfo{}
	}

	return subsystem
}

func GetSystemInfoFromContext(ctx context.Context) SystemInfo {
	systemInfo, ok := ctx.Value(systemInfoContextKey{}).(SystemInfo)
	if !ok {
		return SystemInfo{}
	}

	return systemInfo
}
