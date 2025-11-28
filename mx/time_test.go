package mx_test

import (
	"github.com/morebec/misas/misas"
	"github.com/morebec/misas/mx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRealTimeClock(t *testing.T) {
	t.Run("given non nil timezone, should return date time in this time zone", func(t *testing.T) {
		now := time.Now()
		c := misas.NewRealTimeClock(time.UTC)
		actual := c.Now()
		assertSameTimeWithLeeway(t, now, actual)
		assert.Equal(t, actual.Location(), time.UTC)
	})

	t.Run("given non nil timezone, should return date time in this time zone", func(t *testing.T) {
		now := time.Now()
		c := misas.NewRealTimeClock(nil)
		actual := c.Now()
		assertSameTimeWithLeeway(t, now, c.Now())
		assert.Equal(t, actual.Location(), time.Local)
	})
}

func TestManualClock_Now(t *testing.T) {
	initialDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c := mx.NewManualClock(initialDateTime)
	assert.Equal(t, initialDateTime, c.Now())
}

func TestManualClock_Tick(t *testing.T) {
	initialDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c := mx.NewManualClock(initialDateTime)
	c.Tick(time.Hour)
	assert.Equal(t, lo.Must(time.Parse(time.RFC3339, "2024-01-01T11:00:00Z")), c.Now())
}

func TestManualClock_Set(t *testing.T) {
	initialDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c := mx.NewManualClock(initialDateTime)
	newDateTime := lo.Must(time.Parse(time.RFC3339, "2024-01-15T00:00:00Z"))
	c.Set(newDateTime)
	assert.Equal(t, newDateTime, c.Now())
}

func assertSameTimeWithLeeway(t *testing.T, expected, actual time.Time) {
	leeway := time.Millisecond * 2
	maxExpected := expected.Add(leeway)
	between := isBetween(actual, expected, maxExpected)
	assert.True(t, between)
}

func isBetween(value, min, max time.Time) bool {
	return (value.Equal(min) || value.After(min)) && (value.Equal(max) || value.Before(max))
}

func TestDynamicBindingClock_Swap(t *testing.T) {
	t1 := lo.Must(time.Parse(time.RFC3339, "2024-01-01T10:00:00Z"))
	c1 := mx.NewManualClock(t1)

	t2 := lo.Must(time.Parse(time.RFC3339, "2000-01-01T10:00:00Z"))
	c2 := mx.NewManualClock(t2)

	hc := mx.NewDynamicBindingClock()
	hc.Bind(c1)
	assert.Equal(t, t1, hc.Now())

	hc.Bind(c2)

	assert.Equal(t, t2, hc.Now())
}
