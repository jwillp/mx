package mx_test

import (
	"sync"
	"testing"

	"github.com/morebec/misas/mx"
	"github.com/stretchr/testify/require"
)

func TestLateBinding_Get(t *testing.T) {
	t.Run("GIVEN not bound WHEN getting THEN panic", func(t *testing.T) {
		b := mx.NewLateBinding[int]()
		require.Panics(t, func() { _ = b.Get() })
	})
	t.Run("GIVEN bound WHEN getting THEN shoudl reutrn value", func(t *testing.T) {
		b := mx.NewLateBinding[int]()
		b.Bind(10)
		v := b.Get()
		require.Equal(t, 10, v)
	})
}

func TestLateBinding_Bind(t *testing.T) {
	t.Run("GIVEN not bound WHEN binding THEN should bind successfully", func(t *testing.T) {
		b := mx.NewLateBinding[string]()
		b.Bind("hello")
		v := b.Get()
		require.Equal(t, "hello", v)
	})
	t.Run("GIVEN already bound WHEN binding THEN should panic", func(t *testing.T) {
		b := mx.NewLateBinding[string]()
		b.Bind("first")
		require.Panics(t, func() { b.Bind("second") })
	})
}

func TestLateBinding_IsBound(t *testing.T) {
	t.Run("GIVEN not bound WHEN IsBound is called THEN returns false", func(t *testing.T) {
		b := mx.NewLateBinding[int]()
		require.False(t, b.IsBound())
	})
	t.Run("GIVEN bound WHEN IsBound is called THEN returns true", func(t *testing.T) {
		b := mx.NewLateBinding[int]()
		b.Bind(42)
		require.True(t, b.IsBound())
	})
}

func TestDynamicBinding_Get(t *testing.T) {
	t.Run("GIVEN not bound WHEN getting THEN panic", func(t *testing.T) {
		b := mx.NewDynamicBinding[int]()
		require.Panics(t, func() { _ = b.Get() })
	})
	t.Run("GIVEN bound WHEN getting THEN should return value", func(t *testing.T) {
		b := mx.NewDynamicBinding[int]()
		b.Bind(20)
		v := b.Get()
		require.Equal(t, 20, v)
	})
	t.Run("GIVEN rebound WHEN getting THEN should return latest value", func(t *testing.T) {
		b := mx.NewDynamicBinding[int]()
		b.Bind(1)
		b.Bind(2)
		v := b.Get()
		require.Equal(t, 2, v)
	})
}

func TestDynamicBinding_Bind(t *testing.T) {
	t.Run("GIVEN not bound WHEN binding THEN should bind successfully", func(t *testing.T) {
		b := mx.NewDynamicBinding[string]()
		b.Bind("hello")
		require.Equal(t, "hello", b.Get())
	})
	t.Run("GIVEN already bound WHEN binding THEN should rebind successfully", func(t *testing.T) {
		b := mx.NewDynamicBinding[string]()
		b.Bind("first")
		b.Bind("second")
		require.Equal(t, "second", b.Get())
	})
}

func TestDynamicBinding_IsBound(t *testing.T) {
	t.Run("GIVEN not bound WHEN IsBound is called THEN returns false", func(t *testing.T) {
		b := mx.NewDynamicBinding[int]()
		require.False(t, b.IsBound())
	})
	t.Run("GIVEN bound WHEN IsBound is called THEN returns true", func(t *testing.T) {
		b := mx.NewDynamicBinding[int]()
		b.Bind(42)
		require.True(t, b.IsBound())
	})
}

func TestDynamicBinding_Concurrency(t *testing.T) {
	t.Run("GIVEN concurrent sets and gets WHEN operations run THEN no data race and last value is readable", func(t *testing.T) {
		b := mx.NewDynamicBinding[int]()
		// ensure there's an initial bound value so readers don't panic
		b.Bind(0)
		const n = 100
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				b.Bind(i)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				_ = b.Get()
			}
		}()
		wg.Wait()
		// last value should be n-1
		v := b.Get()
		require.GreaterOrEqual(t, v, 0)
		require.LessOrEqual(t, v, n-1)
	})
}
