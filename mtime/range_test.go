package mtime_test

import (
	"github.com/morebec/go-misas/mtime"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTimeRange_IsWithin(t *testing.T) {
	type given struct {
		Start time.Time
		End   time.Time
	}

	tests := []struct {
		name  string
		given given
		when  time.Time
		then  bool
	}{
		{
			name: "WHEN time is within range THEN should return true",
			given: given{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-05T00:00:00Z")),
			},
			when: lo.Must(time.Parse(time.RFC3339, "2020-01-04T00:00:00Z")),
			then: true,
		},
		{
			name: "WHEN time is not within range THEN should return false",
			given: given{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-05T00:00:00Z")),
			},
			when: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mtime.TimeRange{
				Start: tt.given.Start,
				End:   tt.given.End,
			}
			if got := r.IsWithin(tt.when); got != tt.then {
				t.Errorf("IsWithin() = %v, want %v", got, tt.then)
			}
		})
	}
}

func TestTimeRange_Overlaps(t *testing.T) {
	tests := []struct {
		name  string
		given mtime.TimeRange
		when  mtime.TimeRange
		then  bool
	}{
		{
			name: "WHEN overlaps THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-15T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-09T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-12T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "WHEN case 1 THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-05T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-20T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "WHEN case 2 THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-20T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-19T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-23T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "WHEN case 3 THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-20T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-20T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "WHEN case 4 THEN should return false",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-11T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-20T00:00:00Z")),
			},
			then: false,
		},
		{
			name: "WHEN case 4 THEN should return false",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-19T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-25T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.given.Overlaps(tt.when); got != tt.then {
				t.Errorf("Overlaps() = %v, want %v", got, tt.then)
			}
		})
	}
}

func TestTimeRange_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		given mtime.TimeRange
		then  bool
	}{
		{
			name: "GIVEN start is before end THEN should return false",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-05T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "GIVEN start is equal to end THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "GIVEN start is after end THEN should return false",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-05T00:00:00Z")),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.given.IsValid(); got != tt.then {
				t.Errorf("IsValid() = %v, want %v", got, tt.then)
			}
		})
	}
}

func TestTimeRange_Contains(t *testing.T) {
	tests := []struct {
		name  string
		given mtime.TimeRange
		when  mtime.TimeRange
		then  bool
	}{
		{
			name: "WHEN other time range is between time range THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-15T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-12T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-13T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "WHEN other time range is equal to time range THEN should return true",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-15T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-12T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-13T00:00:00Z")),
			},
			then: true,
		},
		{
			name: "WHEN other time range starts before time range THEN should return false",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-15T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-15T00:00:00Z")),
			},
			then: false,
		},
		{
			name: "WHEN other time range ends after time range THEN should return false",
			given: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-15T00:00:00Z")),
			},
			when: mtime.TimeRange{
				Start: lo.Must(time.Parse(time.RFC3339, "2020-01-12T00:00:00Z")),
				End:   lo.Must(time.Parse(time.RFC3339, "2020-01-16T00:00:00Z")),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.given.Contains(tt.when); got != tt.then {
				t.Errorf("Contains() = %v, want %v", got, tt.then)
			}
		})
	}
}

func TestNewTimeRange(t *testing.T) {
	type args struct {
		start time.Time
		end   time.Time
	}
	type expect struct {
		timeRange mtime.TimeRange
		err       require.ErrorAssertionFunc
	}
	tests := []struct {
		name string
		when args
		then expect
	}{
		{
			name: "WHEN start is before end THEN should return no error",
			when: args{
				start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				end:   lo.Must(time.Parse(time.RFC3339, "2020-01-12T00:00:00Z")),
			},
			then: expect{
				timeRange: mtime.TimeRange{
					Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
					End:   lo.Must(time.Parse(time.RFC3339, "2020-01-12T00:00:00Z")),
				},
				err: require.NoError,
			},
		},
		{
			name: "WHEN start is equal to end THEN should return no error",
			when: args{
				start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				end:   lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
			},
			then: expect{
				timeRange: mtime.TimeRange{
					Start: lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
					End:   lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
				},
				err: require.NoError,
			},
		},
		{
			name: "GIVEN end is before start THEN should return error",
			when: args{
				start: lo.Must(time.Parse(time.RFC3339, "2020-01-10T00:00:00Z")),
				end:   lo.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")),
			},
			then: expect{
				err: require.Error,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mtime.NewTimeRange(tt.when.start, tt.when.end)
			assert.Equal(t, tt.then.timeRange, got)
			tt.then.err(t, err)
		})
	}
}
