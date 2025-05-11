package errors

import (
	"fmt"
)

// FormatOptions defines flags that control how errors and their stack traces
// are rendered in both string and JSON formats.
//
// Fields:
//   - InvertOutput (bool): Flag that inverts the error output (wrap errors shown first).
//   - WithTrace (bool): Flag that enables stack trace output.
//   - InvertTrace (bool): Flag that inverts the stack trace output (top of call stack shown first).
//   - WithExternal (bool): Flag that enables external error output.
type FormatOptions struct {
	InvertOutput bool
	WithTrace    bool
	InvertTrace  bool
	WithExternal bool
}

// UnpackedError contains all decomposed pieces of an error: the root error message
// and stack (ErrRoot), the chain of wrapping errors (ErrChain), and any external
// error not created by this package (ErrExternal).
//
// Fields:
//   - ErrExternal (error): The first non-package error encountered.
//   - ErrRoot (ErrRoot): The original root error and its stack trace.
//   - ErrChain ([]ErrLink): Messages and frames for each wrapping error.
type UnpackedError struct {
	ErrExternal error
	ErrRoot     ErrRoot
	ErrChain    []ErrLink
}

// ErrRoot holds the message and full stack trace for the root error.
//
// Fields:
//   - Msg (string): The root error message.
//   - Stack (Stack): Captured stack frames of the root.
type ErrRoot struct {
	Msg   string
	Stack Stack
}

// formatStr serializes the root error to a string using StringFormat rules:
//   - Always writes the message, followed by MsgStackSep.
//   - If WithTrace, iterates Stack.format and prefixes each frame.
func (err *ErrRoot) formatStr(format StringFormat) string {
	str := err.Msg + format.MsgStackSep

	if format.Options.WithTrace {
		stackArr := err.Stack.format(format.StackElemSep, format.Options.InvertTrace)

		for i, frame := range stackArr {
			str += format.PreStackSep + frame

			if i < len(stackArr)-1 {
				str += format.ErrorSep
			}
		}
	}

	return str
}

// formatJSON constructs a JSON-ready map for the root error:
//
//	{ "message": Msg, "stack": [ ... ] } if WithTrace, or just { "message": Msg }.
func (err *ErrRoot) formatJSON(format JSONFormat) map[string]interface{} {
	rootMap := make(map[string]interface{})

	rootMap["message"] = err.Msg

	if format.Options.WithTrace {
		rootMap["stack"] = err.Stack.format(format.StackElemSep, format.Options.InvertTrace)
	}

	return rootMap
}

// ErrLink represents one level in the wrap chain: a message and the frame
// at which the wrap occurred.
//
// Fields:
//   - Msg (string): Wrapping error message.
//   - Frame (StackFrame): Single captured stack frame of the wrap.
type ErrLink struct {
	Msg   string
	Frame StackFrame
}

// formatStr serializes a wrap link to a string:
//
//	Msg + MsgStackSep + (PreStackSep + Frame.format) if WithTrace.
func (eLink *ErrLink) formatStr(format StringFormat) string {
	str := eLink.Msg + format.MsgStackSep

	if format.Options.WithTrace {
		str += format.PreStackSep + eLink.Frame.format(format.StackElemSep)
	}

	return str
}

// formatJSON constructs a JSON-ready map for a wrap link:
//
//	{ "message": Msg, "stack": Frame.format } if WithTrace, else just message.
func (eLink *ErrLink) formatJSON(format JSONFormat) map[string]interface{} {
	wrapMap := make(map[string]interface{})

	wrapMap["message"] = eLink.Msg

	if format.Options.WithTrace {
		wrapMap["stack"] = eLink.Frame.format(format.StackElemSep)
	}

	return wrapMap
}

// Unpack traverses any error chain by repeatedly calling Unwrap(err),
// collecting wrapped messages and stack frames into an UnpackedError.
// If a non-wrapped error is encountered, it is recorded as ErrExternal.
func Unpack(err error) UnpackedError {
	var upErr UnpackedError

	for err != nil {
		switch err := err.(type) {
		case *root:
			upErr.ErrRoot.Msg = err.message
			upErr.ErrRoot.Stack = err.stack.get()
		case *wrapped:
			// prepend links in stack trace order
			link := ErrLink{Msg: err.message}

			link.Frame = err.frame.get()

			upErr.ErrChain = append([]ErrLink{link}, upErr.ErrChain...)
		default:
			upErr.ErrExternal = err

			return upErr
		}

		err = Unwrap(err)
	}

	return upErr
}

// formatExternalStr renders an external error either by calling fmt.Sprintf("%+v")
// to include its own stack (if withTrace=true) or fmt.Sprint otherwise.
func formatExternalStr(err error, withTrace bool) string {
	if withTrace {
		return fmt.Sprintf("%+v", err)
	}

	return fmt.Sprint(err)
}
