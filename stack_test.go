package errors

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackFrame_format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		frame     StackFrame
		separator string
		expected  string
	}{
		{
			name: "basic format",
			frame: StackFrame{
				Name: "functionName",
				File: "/path/to/file.go",
				Line: 42,
			},
			separator: "|",
			expected:  "functionName|/path/to/file.go|42",
		},
		{
			name: "empty values",
			frame: StackFrame{
				Name: "",
				File: "",
				Line: 0,
			},
			separator: ",",
			expected:  ",,0",
		},
		{
			name: "special characters in name",
			frame: StackFrame{
				Name: "pkg.(*Type).Method",
				File: "/path with spaces/file.go",
				Line: 100,
			},
			separator: " ",
			expected:  "pkg.(*Type).Method /path with spaces/file.go 100",
		},
		{
			name: "separator with special chars",
			frame: StackFrame{
				Name: "func",
				File: "file.go",
				Line: 1,
			},
			separator: "\t-\t",
			expected:  "func\t-\tfile.go\t-\t1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.frame.format(tt.separator)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStack_format(t *testing.T) {
	t.Parallel()

	frames := Stack{
		{Name: "func1", File: "file1.go", Line: 1},
		{Name: "func2", File: "file2.go", Line: 2},
		{Name: "func3", File: "file3.go", Line: 3},
	}

	tests := []struct {
		name      string
		stack     Stack
		separator string
		invert    bool
		expected  []string
	}{
		{
			name:      "natural order",
			stack:     frames,
			separator: " ",
			invert:    false,
			expected: []string{
				"func3 file3.go 3",
				"func2 file2.go 2",
				"func1 file1.go 1",
			},
		},
		{
			name:      "reverse order",
			stack:     frames,
			separator: "\t",
			invert:    true,
			expected: []string{
				"func1\tfile1.go\t1",
				"func2\tfile2.go\t2",
				"func3\tfile3.go\t3",
			},
		},
		{
			name:      "empty stack",
			stack:     Stack{},
			separator: "|",
			invert:    false,
			expected:  []string{},
		},
		{
			name:      "single frame natural",
			stack:     Stack{{Name: "func", File: "file.go", Line: 1}},
			separator: ",",
			invert:    false,
			expected:  []string{"func,file.go,1"},
		},
		{
			name:      "single frame reverse",
			stack:     Stack{{Name: "func", File: "file.go", Line: 1}},
			separator: ",",
			invert:    true,
			expected:  []string{"func,file.go,1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.stack.format(tt.separator, tt.invert)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFrame_pc(t *testing.T) {
	t.Parallel()

	pc := [1]uintptr{}

	runtime.Callers(1, pc[:])

	f := frame(pc[0])

	result := f.pc()

	assert.Equal(t, pc[0]-1, result, "PC should be decremented by 1")
}

func TestFrame_resolveToStackFrame(t *testing.T) {
	t.Parallel()

	pc := [1]uintptr{}

	runtime.Callers(1, pc[:])

	f := frame(pc[0])

	frames := runtime.CallersFrames([]uintptr{pc[0] - 1})

	runtimeFrame, _ := frames.Next()

	expectedName := runtimeFrame.Function

	if idx := strings.LastIndex(expectedName, "/"); idx >= 0 {
		expectedName = expectedName[idx+1:]
	}

	result := f.resolveToStackFrame()

	assert.Equal(t, expectedName, result.Name)
	assert.Equal(t, runtimeFrame.File, result.File)
	assert.Equal(t, runtimeFrame.Line, result.Line)
}

func TestStack_resolveToStackFrames(t *testing.T) {
	t.Parallel()

	const depth = 3

	var pcs [depth]uintptr

	n := runtime.Callers(1, pcs[:])

	s := stack(pcs[:n])

	result := s.resolveToStackFrames()

	frames := runtime.CallersFrames(pcs[:n])

	for i := 0; ; i++ {
		runtimeFrame, more := frames.Next()

		require.Less(t, i, len(result), "More frames from runtime than from resolveToStackFrames")

		assert.Equal(t, filepath.Base(runtimeFrame.Function), filepath.Base(result[i].Name))
		assert.Equal(t, runtimeFrame.File, result[i].File)
		assert.Equal(t, runtimeFrame.Line, result[i].Line)

		if !more {
			break
		}
	}

	assert.Len(t, result, n, "Should have same number of frames as input PCs")
}

func TestStack_insertPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original stack
		insert   stack
		expected stack
	}{
		{
			name:     "insert single PC to empty",
			original: stack{},
			insert:   stack{0x123},
			expected: stack{0x123},
		},
		{
			name:     "insert single PC to non-empty",
			original: stack{0x111, 0x222},
			insert:   stack{0x333},
			expected: stack{0x111, 0x222, 0x333},
		},
		{
			name:     "insert two PCs with match",
			original: stack{0x111, 0x222, 0x444},
			insert:   stack{0x333, 0x444},
			expected: stack{0x111, 0x222, 0x333, 0x444},
		},
		{
			name:     "insert two PCs without match",
			original: stack{0x111, 0x222},
			insert:   stack{0x333, 0x444},
			expected: stack{0x111, 0x222},
		},
		{
			name:     "insert empty PCs",
			original: stack{0x111},
			insert:   stack{},
			expected: stack{0x111},
		},
		{
			name:     "insert multiple matches",
			original: stack{0x111, 0x444, 0x222, 0x444},
			insert:   stack{0x333, 0x444},
			expected: stack{0x111, 0x333, 0x444, 0x222, 0x444},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := make(stack, len(tt.original))

			copy(s, tt.original)

			s.insertPC(tt.insert)

			assert.Equal(t, tt.expected, s)
		})
	}
}

func TestStack_isGlobal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pcs      stack
		expected bool
	}{
		{
			name:     "empty stack",
			pcs:      stack{},
			expected: false,
		},
		{
			name:     "non-init stack",
			pcs:      stack{0x123, 0x456},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.pcs.isGlobal()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCaller(t *testing.T) {
	t.Parallel()

	_, file, line, ok := runtime.Caller(0)

	require.True(t, ok, "runtime.Caller failed")

	result := caller(1)

	require.NotNil(t, result, "caller returned nil")

	resolved := result.resolveToStackFrame()

	assert.Equal(t, "TestCaller", strings.Split(resolved.Name, ".")[1])
	assert.Equal(t, file, resolved.File)
	assert.Greater(t, resolved.Line, line)

	nilResult := caller(1000)

	assert.Nil(t, nilResult)
}

func TestCallers(t *testing.T) {
	t.Parallel()

	result := callers(0)

	require.NotNil(t, result)

	assert.NotEmpty(t, *result, "Should get at least one frame")

	frames := result.resolveToStackFrames()

	for _, frame := range frames {
		assert.False(t, strings.HasPrefix(frame.Name, "runtime."), "Should filter out runtime frames")
	}

	innerResult := callers(2)

	require.NotNil(t, innerResult)

	innerFrames := innerResult.resolveToStackFrames()

	if len(frames) > 0 && len(innerFrames) > 0 {
		assert.NotEqual(t, frames[0].Name, innerFrames[0].Name, "First frame should be different when skipping")
	}

	emptyResult := callers(1000)

	assert.Empty(t, *emptyResult)
}

func TestInsertHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    stack
		insert   uintptr
		at       int
		expected stack
	}{
		{
			name:     "insert at beginning",
			input:    stack{0x2, 0x3},
			insert:   0x1,
			at:       0,
			expected: stack{0x1, 0x2, 0x3},
		},
		{
			name:     "insert in middle",
			input:    stack{0x1, 0x3},
			insert:   0x2,
			at:       1,
			expected: stack{0x1, 0x2, 0x3},
		},
		{
			name:     "insert at end",
			input:    stack{0x1, 0x2},
			insert:   0x3,
			at:       2,
			expected: stack{0x1, 0x2, 0x3},
		},
		{
			name:     "insert into empty",
			input:    stack{},
			insert:   0x1,
			at:       0,
			expected: stack{0x1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := insert(tt.input, tt.insert, tt.at)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFrame_resolveToStackFrame_EdgeCases(t *testing.T) {
	t.Parallel()

	f := frame(0)

	result := f.resolveToStackFrame()

	assert.Empty(t, result.Name)
	assert.Empty(t, result.File)
	assert.Zero(t, result.Line)
}

func TestStack_resolveToStackFrames_InvalidPCs(t *testing.T) {
	t.Parallel()

	s := stack{0}

	result := s.resolveToStackFrames()

	assert.Len(t, result, 1)
	assert.Empty(t, result[0].Name)
	assert.Empty(t, result[0].File)
	assert.Zero(t, result[0].Line)
}
