package mtime

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidTimeRange = errors.New("invalid time range")

// TimeRange represents a range of time.Time with a Start and an End.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

func NewTimeRange(start time.Time, end time.Time) (TimeRange, error) {
	tr := TimeRange{Start: start, End: end}
	if !tr.IsValid() {
		return TimeRange{}, fmt.Errorf("%w: %s", ErrInvalidTimeRange, "start must be before end")
	}
	return tr, nil
}

func (r TimeRange) IsValid() bool { return IsBeforeOrEqual(r.Start, r.End) }

// Contains indicates if this date range fully contains another one.
func (r TimeRange) Contains(o TimeRange) bool {
	return IsBeforeOrEqual(r.Start, o.Start) && IsAfterOrEqual(r.End, o.End)
}

// IsWithin indicates if a given time.Time is within this range or not.
func (r TimeRange) IsWithin(t time.Time) bool { return IsBetween(t, r.Start, r.End) }

// Overlaps indicates if this TimeRange overlaps with another one.
func (r TimeRange) Overlaps(o TimeRange) bool {
	return IsBeforeOrEqual(r.Start, o.End) && IsAfterOrEqual(r.End, o.Start)
}
