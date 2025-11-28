package mx

import (
	"context"
	"sync/atomic"
)

// hotSwappablePluginManager allows swapping the underlying SystemPluginManager at runtime.
// It is safe for concurrent use.
type hotSwappablePluginManager struct {
	pm atomic.Value // holds SystemPluginManager
}

func newHotSwappablePluginManager(pm SystemPluginManager) *hotSwappablePluginManager {
	h := &hotSwappablePluginManager{}
	if pm != nil {
		h.pm.Store(pm)
	}
	return h
}

func (h *hotSwappablePluginManager) get() SystemPluginManager {
	v := h.pm.Load()
	return v.(SystemPluginManager)
}

func (h *hotSwappablePluginManager) AddPlugin(ctx context.Context, p SystemPlugin) {
	h.get().AddPlugin(ctx, p)
}

func (h *hotSwappablePluginManager) DispatchHook(ctx context.Context, hook SystemPluginHook) {
	h.get().DispatchHook(ctx, hook)
}

func (h *hotSwappablePluginManager) Swap(pm SystemPluginManager) {
	h.pm.Store(pm)
}
