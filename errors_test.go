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
		assert.NotEmpty(t, err.(*root).trace)
	})

	t.Run("with type", func(t *testing.T) {
		t.Parallel()

		err := New("typed error", WithType("TEST_TYPE"))

		require.Error(t, err)
		assert.Equal(t, Type("TEST_TYPE"), err.(*root).errType)
	})

	t.Run("with field", func(t *testing.T) {
		t.Parallel()

		err := New("field error", WithField("key", "value"))

		require.Error(t, err)
		assert.Equal(t, map[string]interface{}{"key": "value"}, err.(*root).fields)
	})

	t.Run("error message", func(t *testing.T) {
		t.Parallel()

		err := New("base error")

		wrappedErr := Wrap(err, "wrapper")

		assert.Equal(t, "wrapper: base error", wrappedErr.Error())
	})

	t.Run("is comparison", func(t *testing.T) {
		t.Parallel()

		err1 := New("error", WithType("TYPE"))
		err2 := New("error", WithType("TYPE"))
		err3 := New("different error")

		assert.True(t, err1.(*root).Is(err2))
		assert.False(t, err1.(*root).Is(err3))
	})

	t.Run("as type assertion", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		var target *root

		assert.True(t, err.(*root).As(&target))
		assert.Equal(t, err, target)
	})

	t.Run("unwrap", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base")
		wrappedErr := Wrap(baseErr, "wrapper")

		assert.Equal(t, baseErr, wrappedErr.(*wrapped).Unwrap())
	})

	t.Run("stack frames", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		pcs := err.(*root).StackFrames()

		assert.NotEmpty(t, pcs)
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var nilErr *root

		assert.Equal(t, "<nil>", nilErr.Error())
		assert.Nil(t, nilErr.Fields())
		assert.Empty(t, nilErr.StackFrames())
		assert.False(t, nilErr.Is(errors.New("test")))
	})
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	t.Run("basic wrapping", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base")
		wrappedErr := Wrap(baseErr, "wrapper")

		require.Error(t, wrappedErr)
		assert.Equal(t, "wrapper: base", wrappedErr.Error())
	})

	t.Run("wrapping non-package error", func(t *testing.T) {
		t.Parallel()

		stdErr := errors.New("standard error")
		wrappedErr := Wrap(stdErr, "wrapper")

		require.Error(t, wrappedErr)
		assert.Equal(t, "wrapper: standard error", wrappedErr.Error())
	})

	t.Run("stack frames", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base")
		wrappedErr := Wrap(baseErr, "wrapper")

		pcs := wrappedErr.(*wrapped).StackFrames()

		assert.Len(t, pcs, 1)
	})

	t.Run("preserves root stack", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base")
		wrappedErr1 := Wrap(baseErr, "wrapper1")
		wrappedErr2 := Wrap(wrappedErr1, "wrapper2")

		rootErr := Cause(wrappedErr2).(*root)

		assert.NotEmpty(t, rootErr.trace)
	})

	t.Run("double wrapping with fields", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base", WithField("base_key", "base_value"))
		wrappedErr := Wrap(baseErr, "wrapper", WithField("wrap_key", "wrap_value"))

		assert.Equal(t, map[string]interface{}{"base_key": "base_value"}, baseErr.(*root).fields)
		assert.Equal(t, map[string]interface{}{"wrap_key": "wrap_value"}, wrappedErr.(*wrapped).fields)
	})
}

func TestErrorOptions(t *testing.T) {
	t.Parallel()

	t.Run("with type", func(t *testing.T) {
		t.Parallel()

		opt := WithType("TEST")
		err := New("error", opt)

		assert.Equal(t, Type("TEST"), err.(*root).errType)
	})

	t.Run("with field", func(t *testing.T) {
		t.Parallel()

		opt := WithField("key", "value")
		err := New("error", opt)

		assert.Equal(t, map[string]interface{}{"key": "value"}, err.(*root).fields)
	})

	t.Run("multiple options", func(t *testing.T) {
		t.Parallel()

		err := New("error",
			WithType("TYPE"),
			WithField("key1", "value1"),
			WithField("key2", "value2"),
		)

		assert.Equal(t, Type("TYPE"), err.(*root).errType)
		assert.Equal(t, map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}, err.(*root).fields)
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

		var target *root

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

	t.Run("join nil errors", func(t *testing.T) {
		t.Parallel()

		err := Join(nil, nil)

		assert.NoError(t, err)
	})
}

func TestGlobalError(t *testing.T) {
	t.Parallel()

	t.Run("global flag set", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		assert.False(t, err.(*root).isGlobal)
	})
}

func TestStackPreservation(t *testing.T) {
	t.Parallel()

	t.Run("wrapping preserves original stack", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base")
		wrappedErr := Wrap(baseErr, "wrapper")

		rootErr := Cause(wrappedErr).(*root)

		assert.NotEmpty(t, rootErr.trace)
		assert.Greater(t, len(*rootErr.trace), 1)
	})

	t.Run("double wrapping", func(t *testing.T) {
		t.Parallel()

		baseErr := New("base")
		wrappedErr1 := Wrap(baseErr, "wrapper1")
		wrappedErr2 := Wrap(wrappedErr1, "wrapper2")

		rootErr := Cause(wrappedErr2).(*root)

		assert.NotEmpty(t, rootErr.trace)
	})
}

func TestIs(t *testing.T) {
	t.Parallel()

	err1 := New("1")
	err1a := Wrap(err1, "wrap 2")
	err1b := Wrap(err1a, "wrap 3")

	err2 := errors.New("2")
	err2a := fmt.Errorf("wrap 2: %w", err1)

	joinedErr := Join(err1, err2)

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
		{joinedErr, err1, true},
		{joinedErr, err2, true},
		{joinedErr, New("3"), false},
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

	var r *root

	var w *wrapped

	err1 := New("1")
	err1a := Wrap(err1, "wrap 2")

	joinedErr := Join(err1, err1a)

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
		{joinedErr, &target, true},
		{joinedErr, &r, true},
		{joinedErr, &w, true},
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

func TestJoin(t *testing.T) {
	t.Parallel()

	t.Run("basic join", func(t *testing.T) {
		t.Parallel()

		err1 := New("error1")
		err2 := New("error2")

		joined := Join(err1, err2)

		require.Error(t, joined)
		assert.Equal(t, "error1\nerror2", joined.Error())
	})

	t.Run("join single error", func(t *testing.T) {
		t.Parallel()

		err := New("error")
		joined := Join(err)

		assert.Equal(t, err, joined)
	})

	t.Run("join no errors", func(t *testing.T) {
		t.Parallel()

		joined := Join()

		assert.NoError(t, joined)
	})

	t.Run("join with nil", func(t *testing.T) {
		t.Parallel()

		err := New("error")

		joined := Join(nil, err, nil)

		assert.Equal(t, err, joined)
	})

	t.Run("joined error is", func(t *testing.T) {
		t.Parallel()

		err1 := New("error1")
		err2 := New("error2")

		joined := Join(err1, err2).(*joined)

		assert.True(t, joined.Is(err1))
		assert.True(t, joined.Is(err2))
		assert.False(t, joined.Is(New("error3")))
	})

	t.Run("joined error as", func(t *testing.T) {
		t.Parallel()

		err1 := New("error1")
		err2 := New("error2")

		joined := Join(err1, err2).(*joined)

		var target *root

		assert.True(t, joined.As(&target))
	})

	t.Run("joined stack frames", func(t *testing.T) {
		t.Parallel()

		err1 := New("error1")
		err2 := New("error2")

		joined := Join(err1, err2).(*joined)

		frames := joined.StackFrames()

		assert.NotEmpty(t, frames)
	})

	t.Run("joined unwrap", func(t *testing.T) {
		t.Parallel()

		err1 := New("error1")
		err2 := New("error2")

		joined := Join(err1, err2).(*joined)

		unwrapped := joined.Unwrap()

		assert.Equal(t, []error{err1, err2}, unwrapped)
	})
}

func TestCause(t *testing.T) {
	t.Parallel()

	t.Run("cause of root", func(t *testing.T) {
		t.Parallel()

		err := New("error")
		cause := Cause(err)

		assert.Equal(t, err, cause)
	})

	t.Run("cause of wrapped", func(t *testing.T) {
		t.Parallel()

		rootErr := New("root")
		wrappedErr := Wrap(rootErr, "wrapped")

		cause := Cause(wrappedErr)

		assert.Equal(t, rootErr, cause)
	})

	t.Run("cause of double wrapped", func(t *testing.T) {
		t.Parallel()

		rootErr := New("root")
		wrappedErr1 := Wrap(rootErr, "wrapped1")
		wrappedErr2 := Wrap(wrappedErr1, "wrapped2")

		cause := Cause(wrappedErr2)

		assert.Equal(t, rootErr, cause)
	})

	t.Run("cause of joined", func(t *testing.T) {
		t.Parallel()

		err1 := New("error1")
		err2 := New("error2")

		joined := Join(err1, err2)

		cause := Cause(joined)

		assert.Equal(t, joined, cause) // Joined error is the root cause
	})
}
