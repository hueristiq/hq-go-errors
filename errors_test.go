package errors

import (
	"errors"
	"fmt"
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

func TestIs(t *testing.T) {
	t.Parallel()

	err1 := New("1")
	err1a := Wrap(err1, "wrap 2")
	err1b := Wrap(err1a, "wrap 3")

	err2 := errors.New("2")
	err2a := fmt.Errorf("wrap 2: %w", err1)

	tests := []struct {
		err    error
		target error
		match  bool
	}{
		{nil, nil, true},
		{nil, err1, false},
		{err1, nil, false},
		{err1, err1, true},
		{err1a, err1, true},
		{err1b, err1, true},
		{nil, err2, false},
		{err2, nil, false},
		{err2, err2, true},
		{err2a, err2, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			if tt.match {
				assert.True(t, Is(tt.err, tt.target))
			} else {
				assert.False(t, Is(tt.err, tt.target))
			}
		})
	}
}

func TestAs(t *testing.T) {
	t.Parallel()

	var target Error

	var r *rootError

	var w *wrapError

	err1 := New("1")
	err1a := Wrap(err1, "wrap 2")

	tests := []struct {
		err    error
		target any
		match  bool
	}{
		{nil, nil, false},
		{err1, nil, false},
		{err1, &target, true},
		{err1, &r, true},
		{err1a, &w, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			if tt.match {
				assert.True(t, As(tt.err, tt.target))
			} else {
				assert.False(t, As(tt.err, tt.target))
			}
		})
	}
}
