package mx_test

import (
	"github.com/morebec/misas/mx"
	"github.com/morebec/misas/mxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMessageRegistry_JSONUnmarshal(t *testing.T) {
	t.Run("Given message not registered, THEN should return an error", func(t *testing.T) {
		result, err := mx.CommandRegistry.UnmarshalFromJSON("not_registered", []byte(`{}`))
		assert.Nil(t, result)
		require.Error(t, err)
	})
	t.Run("Given message registered successfully, then return the message", func(t *testing.T) {
		mx.CommandRegistry.Register("registered", mxtest.MockCommand{})
		result, err := mx.CommandRegistry.UnmarshalFromJSON("registered", []byte(`{}`))
		assert.Equal(t, mxtest.MockCommand{}, result)
		require.NoError(t, err)
	})
}
