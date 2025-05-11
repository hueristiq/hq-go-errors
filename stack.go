package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// stack represents a slice of program counters recorded from the call stack.
// Internally, it captures raw PCs (program counters) so that error-handling code
// can later resolve and format a complete backtrace, making it easier to pinpoint
// failure points.
type stack []uintptr

// get resolves the recorded program counters into a slice of detailed StackFrame objects.
// It iterates over each frame in the call stack, extracts the function name (trimming
// any import path for brevity), source file path, and line number. This enables clients
// to present a clear, ordered trace of calls leading up to an error.
//
// The resolution process:
//  1. Converts raw PCs to runtime.Frame objects using runtime.CallersFrames
//  2. Extracts and simplifies function names by removing package paths
//  3. Constructs StackFrame objects with relevant debug information
//
// Returns:
//   - ([]StackFrame): the detailed, ordered frames representing the captured backtrace,
//     with the most recent call first in the slice.
func (s *stack) get() []StackFrame {
	pcs := *s
	frames := runtime.CallersFrames(pcs)

	out := make([]StackFrame, 0, len(pcs))

	for {
		fr, more := frames.Next()

		name := fr.Function

		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		out = append(out, StackFrame{
			Name: name,
			File: fr.File,
			Line: fr.Line,
		})

		if !more {
			break
		}
	}

	return out
}

// isGlobal checks if the captured call stack includes a global init invocation.
// This is useful to detect whether an error occurred during package initialization
// rather than at runtime business logic. It examines each frame's function name
// looking for runtime initialization markers.
//
// Returns:
//   - (bool): true if the stack originates from a global init function, false otherwise
func (s *stack) isGlobal() bool {
	frames := s.get()

	for _, f := range frames {
		if strings.EqualFold(f.Name, "runtime.doinit") {
			return true
		}
	}

	return false
}

// insertPC integrates additional program counters into the existing stack trace.
// It supports two scenarios:
//   - Single-PC insertion: appends a marker (e.g., an error-wrap point)
//   - Dual-PC insertion: locates the context frame then injects the wrapper frame
//     immediately before it, preserving logical call ordering
//
// The insertion logic:
//  1. For single PC, simply appends to the end of the stack
//  2. For dual PCs, searches for the second PC in the existing stack
//     and inserts the first PC before it if found
//
// Parameters:
//   - wrapPCs (stack): program counters to be merged into the current stack.
func (s *stack) insertPC(wrapPCs stack) {
	if len(wrapPCs) < 1 {
		return
	}

	if len(wrapPCs) == 1 {
		*s = append(*s, wrapPCs[0])

		return
	}

	for i, pc := range *s {
		if pc == wrapPCs[0] {
			return
		}

		if pc == wrapPCs[1] {
			*s = append((*s)[:i], append(stack{wrapPCs[0]}, (*s)[i:]...)...)

			return
		}
	}
}

// Stack represents a high-level, resolved backtrace composed of StackFrame entries.
// It enables formatting and presentation of the full call sequence in a human-readable
// format. The Stack type provides methods for formatting the trace in various ways.
type Stack []StackFrame

// format serializes the Stack into human-readable strings suitable for logging
// or error messages. It allows customization of the separator between frame elements
// and the order of presentation (natural or reversed).
//
// Parameters:
//   - separator (string): delimiter between frame elements (e.g., " " or "\t")
//   - invert (bool): order flag: true for reverse (most recent call last),
//     false for natural (most recent call first)
//
// Returns:
//   - ([]string): formatted lines representing each call frame, ordered according
//     to the invert parameter
func (s Stack) format(separator string, invert bool) []string {
	n := len(s)
	str := make([]string, n)

	for i, f := range s {
		idx := i

		if !invert {
			idx = n - 1 - i
		}

		str[idx] = f.format(separator)
	}

	return str
}

// frame represents a single raw program counter from the call stack.
// It exposes methods to resolve metadata about that call site. The frame type
// is used primarily for capturing individual call sites rather than full traces.
type frame uintptr

// pc computes a valid program counter for runtime lookups by subtracting one
// (per the Go runtime's call-instruction convention). This adjustment is necessary
// because the program counter recorded during function calls is actually the
// next instruction after the call.
//
// Returns:
//   - (uintptr): adjusted PC for retrieving function details from the runtime
func (f frame) pc() uintptr {
	return uintptr(f) - 1
}

// get resolves a single frame into a StackFrame, capturing function name,
// file, and line information. It performs the same name simplification as
// stack.get() for consistency.
//
// Returns:
//   - (StackFrame): enriched metadata for this call site containing:
func (f frame) get() StackFrame {
	pc := f.pc()

	fr, _ := runtime.CallersFrames([]uintptr{pc}).Next()

	name := fr.Function

	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	return StackFrame{
		Name: name,
		File: fr.File,
		Line: fr.Line,
	}
}

// StackFrame holds metadata for a single call site within a backtrace.
// It contains all the information needed to identify and locate the
// source of a function call in the codebase.
//
// Fields:
//   - Name (string): simplified function name (without package path) for concise display
//   - File (string): full path of the source file where the call originated
//   - Line (int): exact line number in the source file where the call occurred
type StackFrame struct {
	Name string
	File string
	Line int
}

// format outputs a single-line representation of the StackFrame using the
// provided separator, ideal for log lines or multi-line error dumps.
// The format is consistent and parsable: "Name<sep>File<sep>Line".
//
// Returns:
//   - (string): formatted frame information as a single string
func (f *StackFrame) format(separator string) string {
	return fmt.Sprintf("%s%s%s%s%d", f.Name, separator, f.File, separator, f.Line)
}

// caller captures the immediate caller's frame, skipping over internal frames.
// This is useful for annotating errors with the exact call site in application code.
// The skip parameter allows control over how many stack frames to ascend.
//
// Parameters:
//   - skip (int): number of additional application frames to skip (0 = direct caller)
//
// Returns:
//   - (*frame): pointer to the resolved frame metadata, or nil if no frames available
func caller(skip int) *frame {
	// Maximum depth of stack to capture
	const callersDepth = 32

	var pcs [callersDepth]uintptr

	// +2 skips:
	//   1. this function (caller)
	//   2. the Callers function itself
	n := runtime.Callers(skip+2, pcs[:])
	if n == 0 {
		return nil
	}

	f := frame(pcs[0])

	return &f
}

// callers captures the full application call stack, filtering out runtime internals.
// It returns a stack object that can be further resolved or formatted. The skip
// parameter allows the caller to omit wrapper functions from the trace.
//
// Parameters:
//   - skip (int): number of initial frames to omit (e.g., error wrapper functions)
//
// Returns:
//   - (*stack): stack of filtered program counters ready for resolution,
//     or empty stack if no frames available
func callers(skip int) *stack {
	// Maximum depth of stack to capture
	const depth = 64

	var pcs [depth]uintptr

	// +1 skips the callers function itself
	n := runtime.Callers(skip+1, pcs[:])
	if n == 0 {
		return &stack{}
	}

	valid := pcs[:n]

	out := make(stack, 0, n)

	// Filter out runtime-related frames
	for _, pc := range valid {
		fn := runtime.FuncForPC(pc)
		if fn == nil || strings.HasPrefix(fn.Name(), "runtime.") {
			continue
		}

		out = append(out, pc)
	}

	return &out
}
