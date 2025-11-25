package misas

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
