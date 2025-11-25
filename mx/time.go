package mx

import (
	"github.com/morebec/misas/misas"
	"sync/atomic"
	"time"
)

// ManualClock implementation of a clock that allows to manually specify the
// values that the clock returns. This implementation's primary use case is in
// tests where greater control of time might be needed.
type ManualClock struct {
	currentDateTime time.Time
}

func NewManualClock(currentDateTime time.Time) *ManualClock {
	return &ManualClock{currentDateTime: currentDateTime}
}

// Tick makes the clock tick by adding a specific duration to its current internal date time.
func (c *ManualClock) Tick(duration time.Duration) {
	c.currentDateTime = c.currentDateTime.Add(duration)
}

func (c *ManualClock) Set(dt time.Time) {
	c.currentDateTime = dt
}

func (c *ManualClock) Now() time.Time {
	return c.currentDateTime
}

// HotSwappableClock is an implementation of a clock that allows to change its
// underlying clock at runtime using atomic operations for concurrency safety.
type HotSwappableClock struct {
	clock atomic.Value
}

func NewHotSwappableClock(clock misas.Clock) *HotSwappableClock {
	hc := &HotSwappableClock{}
	hc.clock.Store(clock)
	return hc
}

func (hc *HotSwappableClock) Now() time.Time         { return hc.clock.Load().(misas.Clock).Now() }
func (hc *HotSwappableClock) Swap(clock misas.Clock) { hc.clock.Store(clock) }
