package errors

import (
	"fmt"
)

// FormatOptions defines flags that control how errors and their stack traces
// are rendered in both string and JSON formats.
//
// Fields:
//   - InvertOutput bool: Flag that inverts the error output (wrap errors shown first).
//   - WithTrace    bool: Flag that enables stack trace output.
//   - InvertTrace  bool: Flag that inverts the stack trace output (top of call stack shown first).
//   - WithExternal bool: Flag that enables external error output.
type FormatOptions struct {
	InvertOutput bool
	WithTrace    bool
	InvertTrace  bool
	WithExternal bool
}

// StringFormat holds formatting rules and separators for rendering errors
// as human-readable strings.
//
// Fields:
//   - Options      FormatOptions: Format options (e.g. omitting stack trace or inverting the output order).
//   - MsgStackSep  string: Separator between error messages and stack frame data.
//   - PreStackSep  string: Separator at the beginning of each stack frame.
//   - StackElemSep string: Separator between elements of each stack frame.
//   - ErrorSep     string: Separator between each error in the chain.
type StringFormat struct {
	Options      FormatOptions
	MsgStackSep  string
	PreStackSep  string
	StackElemSep string
	ErrorSep     string
}

// NewDefaultStringFormat returns a StringFormat pre-configured with sensible
// separators based on whether stack traces are desired.
//
// If options.WithTrace is true, uses:
//
//	MsgStackSep  = "\n"
//	PreStackSep  = "\t"
//	StackElemSep = ":"
//	ErrorSep     = "\n"
//
// Otherwise, uses:
//
//	ErrorSep     = ": "
func NewDefaultStringFormat(options FormatOptions) StringFormat {
	stringFmt := StringFormat{
		Options: options,
	}

	if options.WithTrace {
		stringFmt.MsgStackSep = "\n"
		stringFmt.PreStackSep = "\t"
		stringFmt.StackElemSep = ":"
		stringFmt.ErrorSep = "\n"
	} else {
		stringFmt.ErrorSep = ": "
	}

	return stringFmt
}

// ToString formats any error into a human-readable string.
// It uses a default StringFormat configured with the given withTrace flag
// and always includes any external error.
//
// Without trace (withTrace=false):
//
//	wrapErrorMsg: rootErrorMsg
//
// With trace (withTrace=true):
//
//	wrapErrorMsg
//	   Method2:File2:Line2
//	rootErrorMsg
//	   Method2:File2:Line2
//	   Method1:File1:Line1
func ToString(err error, withTrace bool) string {
	return ToCustomString(err, NewDefaultStringFormat(FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// ToCustomString returns a custom formatted string for a given error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as follows:
//
//	<Wrap error msg>[Format.ErrorSep]<Root error msg>
//
// An error with trace will be formatted as follows:
//
//	<Wrap error msg>[Format.MsgStackSep]
//	[Format.PreStackSep]<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>[Format.ErrorSep]
//	<Root error msg>[Format.MsgStackSep]
//	[Format.PreStackSep]<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>[Format.ErrorSep]
//	[Format.PreStackSep]<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>[Format.ErrorSep]
func ToCustomString(err error, format StringFormat) string {
	upErr := Unpack(err)

	var str string

	if format.Options.InvertOutput {
		if format.Options.WithExternal && upErr.ErrExternal != nil {
			str += formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)

			if (format.Options.WithTrace && len(upErr.ErrRoot.Stack) > 0) || upErr.ErrRoot.Msg != "" {
				str += format.ErrorSep
			}
		}

		str += upErr.ErrRoot.formatStr(format)

		for _, eLink := range upErr.ErrChain {
			str += format.ErrorSep + eLink.formatStr(format)
		}
	} else {
		for i := len(upErr.ErrChain) - 1; i >= 0; i-- {
			str += upErr.ErrChain[i].formatStr(format) + format.ErrorSep
		}

		str += upErr.ErrRoot.formatStr(format)

		if format.Options.WithExternal && upErr.ErrExternal != nil {
			if (format.Options.WithTrace && len(upErr.ErrRoot.Stack) > 0) || upErr.ErrRoot.Msg != "" {
				str += format.ErrorSep
			}

			str += formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
		}
	}

	return str
}

// JSONFormat holds formatting rules for serializing errors into JSON-like maps.
// The StackElemSep separates method, file, and line in each frame string.
//
// Fields:
//   - Options      FormatOptions: Format options (e.g. omitting stack trace or inverting the output order).
//   - StackElemSep string: Separator between elements of each stack frame.
type JSONFormat struct {
	Options      FormatOptions
	StackElemSep string
}

// NewDefaultJSONFormat returns a JSONFormat with default separators.
// StackElemSep defaults to ":".
func NewDefaultJSONFormat(options FormatOptions) JSONFormat {
	return JSONFormat{
		Options:      options,
		StackElemSep: ":",
	}
}

// ToJSON returns a JSON formatted map for a given error.
//
// An error without trace will be formatted as follows:
//
//	{
//	  "root": {
//	      "message": "Root error msg"
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg"
//	    }
//	  ]
//	}
//
// An error with trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	    "stack": [
//	      "<Method2>:<File2>:<Line2>",
//	      "<Method1>:<File1>:<Line1>"
//	    ]
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg",
//	      "stack": "<Method2>:<File2>:<Line2>"
//	    }
//	  ]
//	}
func ToJSON(err error, withTrace bool) map[string]interface{} {
	return ToCustomJSON(err, NewDefaultJSONFormat(FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// ToCustomJSON returns a JSON formatted map for a given error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	  },
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg'",
//	    }
//	  ]
//	}
//
// An error with trace will be formatted as follows:
//
//	{
//	  "root": {
//	    "message": "Root error msg",
//	    "stack": [
//	      "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>",
//	      "<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>"
//	    ]
//	  }
//	  "wrap": [
//	    {
//	      "message": "Wrap error msg",
//	      "stack": "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>"
//	    }
//	  ]
//	}
func ToCustomJSON(err error, format JSONFormat) map[string]interface{} {
	upErr := Unpack(err)

	jsonMap := make(map[string]interface{})

	if format.Options.WithExternal && upErr.ErrExternal != nil {
		jsonMap["external"] = formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
	}

	if upErr.ErrRoot.Msg != "" || len(upErr.ErrRoot.Stack) > 0 {
		jsonMap["root"] = upErr.ErrRoot.formatJSON(format)
	}

	if len(upErr.ErrChain) > 0 {
		var wrapArr []map[string]interface{}

		for _, eLink := range upErr.ErrChain {
			wrapMap := eLink.formatJSON(format)

			if format.Options.InvertOutput {
				wrapArr = append(wrapArr, wrapMap)
			} else {
				wrapArr = append([]map[string]interface{}{wrapMap}, wrapArr...)
			}
		}

		jsonMap["wrap"] = wrapArr
	}

	return jsonMap
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

// UnpackedError contains all decomposed pieces of an error: the root error message
// and stack (ErrRoot), the chain of wrapping errors (ErrChain), and any external
// error not created by this package (ErrExternal).
//
// Fields:
//   - ErrExternal error: The first non-package error encountered.
//   - ErrRoot     ErrRoot: The original root error and its stack trace.
//   - ErrChain    []ErrLink: Messages and frames for each wrapping error.
type UnpackedError struct {
	ErrExternal error
	ErrRoot     ErrRoot
	ErrChain    []ErrLink
}

// formatExternalStr renders an external error either by calling fmt.Sprintf("%+v")
// to include its own stack (if withTrace=true) or fmt.Sprint otherwise.
func formatExternalStr(err error, withTrace bool) string {
	if withTrace {
		return fmt.Sprintf("%+v", err)
	}

	return fmt.Sprint(err)
}

// ErrRoot holds the message and full stack trace for the root error.
//
// Fields:
//   - Msg   string: The root error message.
//   - Stack Stack: Captured stack frames of the root.
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
//   - Msg   string: Wrapping error message.
//   - Frame StackFrame: Single captured stack frame of the wrap.
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
