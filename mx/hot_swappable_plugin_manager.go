package mx

import (
	"context"
	"sync/atomic"
)

// hotSwappablePluginManager allows swapping the underlying PluginManager at runtime.
// It is safe for concurrent use.
type hotSwappablePluginManager struct {
	pm atomic.Value // holds PluginManager
}

func newHotSwappablePluginManager(pm PluginManager) *hotSwappablePluginManager {
	h := &hotSwappablePluginManager{}
	if pm != nil {
		h.pm.Store(pm)
	}
	return h
}

func (h *hotSwappablePluginManager) get() PluginManager {
	v := h.pm.Load()
	return v.(PluginManager)
}

func (h *hotSwappablePluginManager) AddPlugin(ctx context.Context, p Plugin) {
	h.get().AddPlugin(ctx, p)
}

func (h *hotSwappablePluginManager) DispatchHook(ctx context.Context, hook PluginHook) {
	h.get().DispatchHook(ctx, hook)
}

func (h *hotSwappablePluginManager) Swap(pm PluginManager) {
	h.pm.Store(pm)
}
