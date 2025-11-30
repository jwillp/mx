package mtime_test

import (
	"github.com/morebec/go-misas/mtime"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_IsAfterOrEqual(t *testing.T) {
	type args struct {
		a time.Time
		b time.Time
	}
	now := time.Now()
	tests := []struct {
		name string
		when args
		then bool
	}{
		{
			name: "WHEN both are equal THEN should return true",
			when: args{
				a: now,
				b: now,
			},
			then: true,
		},
		{
			name: "WHEN a is after b THEN should return true",
			when: args{
				a: now.Add(time.Hour),
				b: now,
			},
			then: true,
		},
		{
			name: "WHEN a is before b THEN should return false",
			when: args{
				a: now,
				b: now.Add(time.Hour),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.then, mtime.IsAfterOrEqual(tt.when.a, tt.when.b), "IsAfterOrEqual(%v, %v)", tt.when.a, tt.when.b)
		})
	}
}

func Test_IsBeforeOrEqual(t *testing.T) {
	type args struct {
		a time.Time
		b time.Time
	}
	now := time.Now()
	tests := []struct {
		name string
		when args
		then bool
	}{
		{
			name: "WHEN both are equal THEN should return true",
			when: args{
				a: now,
				b: now,
			},
			then: true,
		},
		{
			name: "WHEN a is before b THEN should return true",
			when: args{
				a: now,
				b: now.Add(time.Hour),
			},
			then: true,
		},
		{
			name: "WHEN a is after b THEN should return false",
			when: args{
				a: now.Add(time.Hour),
				b: now,
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.then,
				mtime.IsBeforeOrEqual(tt.when.a, tt.when.b),
				"IsBeforeOrEqual(%v, %v)",
				tt.when.a,
				tt.when.b,
			)
		})
	}
}

func Test_IsBetween(t *testing.T) {
	type args struct {
		time time.Time
		min  time.Time
		max  time.Time
	}
	now := time.Now()
	tests := []struct {
		name string
		when args
		then bool
	}{
		{
			name: "WHEN time is within min and max THEN should return true",
			when: args{
				time: now.Add(time.Second),
				min:  now,
				max:  now.Add(time.Hour),
			},
			then: true,
		},
		{
			name: "WHEN time is before than min THEN should return false",
			when: args{
				time: now,
				min:  now.Add(time.Second),
				max:  now.Add(time.Hour),
			},
			then: false,
		},
		{
			name: "WHEN time is after max THEN should return false",
			when: args{
				time: now.Add(time.Hour),
				min:  now,
				max:  now.Add(time.Second),
			},
			then: false,
		},
		{
			name: "WHEN time is equal to min THEN should return true",
			when: args{
				time: now,
				min:  now,
				max:  now.Add(time.Hour),
			},
			then: true,
		},
		{
			name: "WHEN time is equal to max THEN should return true",
			when: args{
				time: now.Add(time.Hour),
				min:  now,
				max:  now.Add(time.Hour),
			},
			then: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.then,
				mtime.IsBetween(tt.when.time, tt.when.min, tt.when.max),
				"IsBetween(%v, %v, %v)",
				tt.when.time,
				tt.when.min,
				tt.when.max,
			)
		})
	}
}

func Test_IsBetweenExclusive(t *testing.T) {
	type args struct {
		time time.Time
		min  time.Time
		max  time.Time
	}
	now := time.Now()
	tests := []struct {
		name string
		when args
		then bool
	}{
		{
			name: "WHEN time is within min and max THEN should return true",
			when: args{
				time: now.Add(time.Second),
				min:  now,
				max:  now.Add(time.Hour),
			},
			then: true,
		},
		{
			name: "WHEN time is before than min THEN should return false",
			when: args{
				time: now,
				min:  now.Add(time.Second),
				max:  now.Add(time.Hour),
			},
			then: false,
		},
		{
			name: "WHEN time is after max THEN should return false",
			when: args{
				time: now.Add(time.Hour),
				min:  now,
				max:  now.Add(time.Second),
			},
			then: false,
		},
		{
			name: "WHEN time is equal to min THEN should return false",
			when: args{
				time: now,
				min:  now,
				max:  now.Add(time.Hour),
			},
			then: false,
		},
		{
			name: "WHEN time is equal to max THEN should return false",
			when: args{
				time: now.Add(time.Hour),
				min:  now,
				max:  now.Add(time.Hour),
			},
			then: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.then,
				mtime.IsBetweenExclusive(tt.when.time, tt.when.min, tt.when.max),
				"IsBetween(%v, %v, %v)",
				tt.when.time,
				tt.when.min,
				tt.when.max,
			)
		})
	}
}
