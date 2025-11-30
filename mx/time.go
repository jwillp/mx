package mx

import (
	"github.com/morebec/misas/mtime"
	"time"
)

// DynamicBindingClock is an implementation of a clock that allows to change its
// underlying clock at runtime using atomic operations for concurrency safety.
type DynamicBindingClock struct {
	*DynamicBinding[mtime.Clock]
}

func NewDynamicBindingClock() *DynamicBindingClock {
	return &DynamicBindingClock{
		DynamicBinding: NewDynamicBinding[mtime.Clock](),
	}
}

func (hc *DynamicBindingClock) Now() time.Time { return hc.Get().Now() }
