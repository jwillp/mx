package explo

import "time"

type Clock interface {
	Now() time.Time
}

type ClockFunc func() time.Time

func (f ClockFunc) Now() time.Time { return f() }

// NewRealTimeClock returns the main implementation of a [Clock] that returns the
// time of the running operating system in a given time zone. if nil is passed,
// will default to the local time zone.
func NewRealTimeClock(tz *time.Location) ClockFunc {
	if tz != nil {
		time.Local = tz
	}
	return time.Now
}

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
// underlying clock at runtime.
type HotSwappableClock struct {
	Clock
}

func NewHotSwappableClock(clock Clock) *HotSwappableClock { return &HotSwappableClock{Clock: clock} }

func (hc *HotSwappableClock) Swap(clock Clock) { hc.Clock = clock }
