package mtime

import (
	"time"
)

// IsBetween indicates if a time.Time is between two other time.Time values.
func IsBetween(t time.Time, minVal time.Time, maxVal time.Time) bool {
	return !t.Before(minVal) && !t.After(maxVal) // equivalent to !(t.Before(minVal) || t.After(maxVal))
}

// IsBetweenExclusive indicates if a time.Time is strictly between two other
// time.Time values. This means that if t is equal to minVal or maxVal, it will
// return false.
func IsBetweenExclusive(t time.Time, minVal time.Time, maxVal time.Time) bool {
	return t.After(minVal) && t.Before(maxVal)
}

// IsAfterOrEqual indicates if A is after or equal to B.
func IsAfterOrEqual(a time.Time, b time.Time) bool {
	return a.After(b) || a.Equal(b)
}

// IsBeforeOrEqual indicates if A is before or equal to B.
func IsBeforeOrEqual(a time.Time, b time.Time) bool {
	return a.Before(b) || a.Equal(b)
}
