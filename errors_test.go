package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootError(t *testing.T) {
	t.Parallel()

	t.Run("basic functionality", func(t *testing.T) {
		t.Parallel()

		err := New("test error")

		require.Error(t, err)

		assert.Equal(t, "test error", err.Error())
		assert.NotEmpty(t, err.(*rootError).stack)
	})

	t.Run("with type", func(t *testing.T) {
		t.Parallel()

		err := New("typed error", WithType("TEST_TYPE"))

		require.Error(t, err)

		assert.Equal(t, ErrorType("TEST_TYPE"), err.(*rootError).t)
	})

	t.Run("with field", func(t *testing.T) {
		t.Parallel()

		err := New("field error", WithField("key", "value"))

		require.Error(t, err)

		assert.Equal(t, map[string]interface{}{"key": "value"}, err.(*rootError).fields)
	})

	t.Run("error message", func(t *testing.T) {
		t.Parallel()

		err := New("base error")
		wrapped := Wrap(err, "wrapper")

		assert.Equal(t, "wrapper: base error", wrapped.Error())
	})

	t.Run("is comparison", func(t *testing.T) {
		t.Parallel()

		err1 := New("error", WithType("TYPE"))
		err2 := New("error", WithType("TYPE"))
		err3 := New("different error")

		assert.True(t, err1.(*rootError).Is(err2))
		assert.False(t, err1.(*rootError).Is(err3))
	})

	t.Run("as type assertion", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		var target *rootError

		assert.True(t, err.(*rootError).As(&target))
		assert.Equal(t, err, target)
	})

	t.Run("unwrap", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped := Wrap(base, "wrapper")

		assert.Equal(t, base, wrapped.(*wrapError).Unwrap())
	})

	t.Run("stack frames", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		pcs := err.(*rootError).StackFrames()

		assert.NotEmpty(t, pcs)
	})
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	t.Run("basic wrapping", func(t *testing.T) {
		t.Parallel()

		base := New("base")

		wrapped := Wrap(base, "wrapper")

		require.Error(t, wrapped)

		assert.Equal(t, "wrapper: base", wrapped.Error())
	})

	t.Run("wrapping non-package error", func(t *testing.T) {
		t.Parallel()

		stdErr := errors.New("standard error")
		wrapped := Wrap(stdErr, "wrapper")

		require.Error(t, wrapped)

		assert.Equal(t, "wrapper: standard error", wrapped.Error())
	})

	t.Run("stack frames", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped := Wrap(base, "wrapper")

		pcs := wrapped.(*wrapError).StackFrames()

		assert.Len(t, pcs, 1)
	})

	t.Run("preserves root stack", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped1 := Wrap(base, "wrapper1")
		wrapped2 := Wrap(wrapped1, "wrapper2")

		root := Cause(wrapped2).(*rootError)

		assert.NotEmpty(t, root.stack)
	})
}

func TestHelpers(t *testing.T) {
	t.Parallel()

	t.Run("unwrap", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped := Wrap(base, "wrapper")

		assert.Equal(t, base, Unwrap(wrapped))
	})

	t.Run("is", func(t *testing.T) {
		t.Parallel()

		err0 := errors.New("error")

		assert.True(t, Is(err0, err0))
		assert.True(t, Is(Wrap(err0, "wrapper"), err0))

		err1 := New("error", WithType("TYPE"))

		assert.True(t, Is(err1, err1))
		assert.True(t, Is(Wrap(err1, "wrapper"), err1))
	})

	t.Run("as", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		var target *rootError

		assert.True(t, As(err, &target))
		assert.Equal(t, err, target)
	})

	t.Run("cause", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped := Wrap(base, "wrapper")

		assert.Equal(t, base, Cause(wrapped))
	})
}

func TestErrorOptions(t *testing.T) {
	t.Parallel()

	t.Run("with type", func(t *testing.T) {
		t.Parallel()

		opt := WithType("TEST")
		err := New("error", opt)

		assert.Equal(t, ErrorType("TEST"), err.(*rootError).t)
	})

	t.Run("with field", func(t *testing.T) {
		t.Parallel()

		opt := WithField("key", "value")
		err := New("error", opt)

		assert.Equal(t, map[string]interface{}{"key": "value"}, err.(*rootError).fields)
	})

	t.Run("multiple options", func(t *testing.T) {
		t.Parallel()

		err := New("error",
			WithType("TYPE"),
			WithField("key1", "value1"),
			WithField("key2", "value2"),
		)

		assert.Equal(t, ErrorType("TYPE"), err.(*rootError).t)
		assert.Equal(t, map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}, err.(*rootError).fields)
	})
}

func TestNilHandling(t *testing.T) {
	t.Parallel()

	t.Run("new with empty message", func(t *testing.T) {
		t.Parallel()

		err := New("")

		require.Error(t, err)
	})

	t.Run("wrap nil error", func(t *testing.T) {
		t.Parallel()

		err := Wrap(nil, "wrapper")

		assert.NoError(t, err)
	})

	t.Run("is with nil", func(t *testing.T) {
		t.Parallel()

		assert.True(t, Is(nil, nil))
		assert.False(t, Is(New("error"), nil))
	})

	t.Run("as with nil", func(t *testing.T) {
		t.Parallel()

		var target *rootError

		assert.False(t, As(nil, &target))
		assert.Nil(t, target)
	})

	t.Run("unwrap nil", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, Unwrap(nil))
	})

	t.Run("cause nil", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, Cause(nil))
	})
}

func TestGlobalError(t *testing.T) {
	t.Parallel()

	t.Run("global flag set", func(t *testing.T) {
		t.Parallel()

		// This is hard to test directly since we can't easily trigger global init
		// But we can verify the field exists and is set appropriately
		err := New("error")

		assert.False(t, err.(*rootError).global)
	})
}

func TestStackPreservation(t *testing.T) {
	t.Parallel()

	t.Run("wrapping preserves original stack", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped := Wrap(base, "wrapper")

		root := Cause(wrapped).(*rootError)

		assert.NotEmpty(t, root.stack)
		assert.Greater(t, len(*root.stack), 1)
	})

	t.Run("double wrapping", func(t *testing.T) {
		t.Parallel()

		base := New("base")
		wrapped1 := Wrap(base, "wrapper1")
		wrapped2 := Wrap(wrapped1, "wrapper2")

		root := Cause(wrapped2).(*rootError)

		assert.NotEmpty(t, root.stack)
	})
}
