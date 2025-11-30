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
	ctx = context.WithValue(ctx, systemInfoContextKey{}, s.info)

	return ctx
}

func newSubsystemContext(ctx context.Context, info SubsystemInfo) context.Context {
	ctx = context.WithValue(ctx, subsystemInfoContextKey{}, info)

	return ctx
}

type Context struct {
	context.Context
}

func Ctx(ctx context.Context) Context {
	return Context{ctx}
}

func (c Context) SystemInfo() SystemInfo {
	systemInfo, ok := c.Context.Value(systemInfoContextKey{}).(SystemInfo)
	if !ok {
		return SystemInfo{}
	}

	return systemInfo
}

func (c Context) SubsystemInfo() SubsystemInfo {
	subsystemInfo, ok := c.Context.Value(subsystemInfoContextKey{}).(SubsystemInfo)
	if !ok {
		return SubsystemInfo{}
	}

	return subsystemInfo
}
