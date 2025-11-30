package mx

import (
	"context"
)

type systemLoggerContextKey struct{}

type systemInfoContextKey struct{}
type subsystemInfoContextKey struct{}
type subsystemOriginContextKey struct{}

func newSystemContext(s System) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, systemLoggerContextKey{}, s.logger)
	ctx = context.WithValue(ctx, systemInfoContextKey{}, s.info)

	return ctx
}

func newSubsystemContext(ctx context.Context, info SubsystemInfo) context.Context {
	// Track origin: if a subsystem context already exists, use it as origin
	// otherwise, this is the root subsystem
	var origin SubsystemInfo
	if originInfo, ok := ctx.Value(subsystemInfoContextKey{}).(SubsystemInfo); ok {
		origin = originInfo
	}

	ctx = context.WithValue(ctx, subsystemInfoContextKey{}, info)
	if (origin != SubsystemInfo{}) {
		ctx = context.WithValue(ctx, subsystemOriginContextKey{}, origin)
	}

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

func (c Context) SubsystemOrigin() SubsystemInfo {
	origin, ok := c.Context.Value(subsystemOriginContextKey{}).(SubsystemInfo)
	if !ok {
		return SubsystemInfo{}
	}

	return origin
}
