package mtime_test

import (
	mtime2 "github.com/morebec/go-misas/mtime"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRealTimeClock(t *testing.T) {
	t.Run("given non nil timezone, should return date time in this time zone", func(t *testing.T) {
		now := time.Now()
		c := mtime2.NewRealTimeClock(time.UTC)
		actual := c.Now()
		assertSameTimeWithLeeway(t, now, actual)
		assert.Equal(t, actual.Location(), time.UTC)
	})

	t.Run("given non nil timezone, should return date time in this time zone", func(t *testing.T) {
		now := time.Now()
		c := mtime2.NewRealTimeClock(nil)
		actual := c.Now()
		assertSameTimeWithLeeway(t, now, c.Now())
		assert.Equal(t, actual.Location(), time.Local)
	})
}

func TestManualClock_Now(t *testing.T) {
	initialDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c := mtime2.NewManualClock(initialDateTime)
	assert.Equal(t, initialDateTime, c.Now())
}

func TestManualClock_Tick(t *testing.T) {
	initialDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c := mtime2.NewManualClock(initialDateTime)
	c.Tick(time.Hour)
	assert.Equal(t, lo.Must(time.Parse(time.RFC3339, "2024-01-01T11:00:00Z")), c.Now())
}

func TestManualClock_Set(t *testing.T) {
	initialDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c := mtime2.NewManualClock(initialDateTime)
	newDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-15T00:00:00Z"))
	c.Set(newDateTime)
	assert.Equal(t, newDateTime, c.Now())
}

func assertSameTimeWithLeeway(t *testing.T, expected, actual time.Time) {
	leeway := time.Millisecond * 2
	maxExpected := expected.Add(leeway)
	between := mtime2.IsBetween(actual, expected, maxExpected)
	assert.True(t, between)
}

func TestHotSwappableClock_Swap(t *testing.T) {
	t1 := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c1 := mtime2.NewManualClock(t1)
	t2 := lo.Must(time.Parse(time.RFC3339, "2000-01-01T10:00:00Z"))
	c2 := mtime2.NewManualClock(t2)

	hc := mtime2.NewHotSwappableClock(c1)
	assert.Equal(t, t1, hc.Now())

	hc.Swap(c2)

	assert.Equal(t, t2, hc.Now())
}
