package errors

import (
	"fmt"
)

// UnpackedError contains all decomposed pieces of an error chain.
// It provides structured access to the complete error hierarchy.
//
// Fields:
//   - ErrExternal (error): The first non-package error encountered in the chain.
//     This captures errors from other libraries or the standard library.
//   - ErrRoot (ErrRoot): The original root error created by this package.
//     Contains the core error message and stack trace.
//   - ErrChain ([]ErrLink): Ordered sequence of wrapping operations.
//     Each link represents additional context added through error wrapping.
type UnpackedError struct {
	ErrExternal error
	ErrRoot     ErrRoot
	ErrChain    []ErrLink
}

// ErrRoot holds detailed information about a root error created by this package.
// It represents the initial error in a chain before any wrapping operations.
//
// Fields:
//   - Message (string): Primary human-readable error message.
//   - Type (Type): Optional error classification for programmatic handling.
//   - Fields (map[string]interface{}): Structured key-value pairs providing
//     additional context about the error (e.g., input parameters, state values).
//   - Stack (Stack): Complete call stack captured at error creation time.
//     Each frame contains method, file, and line information.
type ErrRoot struct {
	Message string
	Type    Type
	Fields  map[string]interface{}
	Stack   Stack
}

// formatStr serializes the ErrRoot into a human-readable string using the provided StringFormat.
// format.Options controls inclusion of type, fields, and stack trace ordering.
//
// Parameters:
//   - format (*StringFormat): formatting rules and separators.
//
// Returns:
//   - formatted (string): the formatted error string.
func (err *ErrRoot) formatStr(format *StringFormat) (formatted string) {
	if err.Type != "" {
		formatted += format.TypePrefix + string(err.Type) + format.TypeSuffix + format.TypeMessageSeparator
	}

	formatted += err.Message

	if len(err.Fields) > 0 {
		formatted += format.MessageFieldsSeparator + "Fields:" + format.FieldsTitleEntriesSeparator

		first := true

		for k, v := range err.Fields {
			if !first {
				formatted += format.FieldSep
			}

			formatted += fmt.Sprintf("  %s%s%v", k, format.FieldKVSep, v)

			first = false
		}
	}

	if format.Options.WithTrace {
		formatted += format.FieldsStackSeparator + "Stack:" + format.StackTitleEntriesSeparator

		stackArr := err.Stack.format(format.StackElemSep, format.Options.InvertTrace)

		for i, frame := range stackArr {
			formatted += format.PreStackSep + frame

			if i < len(stackArr)-1 {
				formatted += format.StackEntriesSeparator
			}
		}
	}

	return
}

// formatJSON constructs a JSON-ready map from ErrRoot according to JSONFormat rules.
// format.Options.WithTrace controls whether the stack slice is included.
//
// Parameters:
//   - format (*JSONFormat): formatting rules for JSON serialization.
//
// Returns:
//   - formatted (map[string]interface{}): the JSON-like representation of the root error.
func (err *ErrRoot) formatJSON(format *JSONFormat) (formatted map[string]interface{}) {
	formatted = make(map[string]interface{})

	formatted["message"] = err.Message

	if err.Type != "" {
		formatted["type"] = err.Type
	}

	if len(err.Fields) > 0 {
		formatted["fields"] = err.Fields
	}

	if format.Options.WithTrace {
		formatted["stack"] = err.Stack.format(format.StackElemSep, format.Options.InvertTrace)
	}

	return
}

// ErrLink represents a single wrapping operation in an error chain.
// Each wrap adds contextual information while preserving the original error.
//
// Fields:
//   - Message (string): Contextual message added during wrapping.
//   - Type (Type): Optional classification for this wrap level.
//   - Fields (map[string]interface{}): Additional context specific to this wrap.
//   - Frame (StackFrame): Single stack frame capturing where the wrap occurred.
type ErrLink struct {
	Message string
	Type    Type
	Fields  map[string]interface{}
	Frame   StackFrame
}

// formatStr serializes the ErrLink into a string according to StringFormat rules.
// format.Options controls inclusion of type, fields, and single frame output.
//
// Parameters:
//   - format (*StringFormat): formatting rules and separators.
//
// Returns:
//   - formatted (string): the formatted wrap link string.
func (eLink *ErrLink) formatStr(format *StringFormat) (formatted string) {
	if eLink.Type != "" {
		formatted += format.TypePrefix + string(eLink.Type) + format.TypeSuffix + format.TypeMessageSeparator
	}

	formatted += eLink.Message

	if len(eLink.Fields) > 0 {
		formatted += format.MessageFieldsSeparator + "Fields:" + format.FieldsTitleEntriesSeparator

		first := true

		for k, v := range eLink.Fields {
			if !first {
				formatted += format.FieldSep
			}

			formatted += fmt.Sprintf("  %s%s%v", k, format.FieldKVSep, v)

			first = false
		}
	}

	if format.Options.WithTrace {
		formatted += format.FieldsStackSeparator + "Stack:" + format.StackTitleEntriesSeparator

		formatted += format.PreStackSep + eLink.Frame.format(format.StackElemSep)
	}

	return
}

// formatJSON constructs a JSON-ready map from ErrLink according to JSONFormat rules.
// format.Options.WithTrace controls whether the frame is included.
//
// Parameters:
//   - format (*JSONFormat): formatting rules for JSON serialization.
//
// Returns:
//   - formatted (map[string]interface{}): the map representation of the wrap link.
func (eLink *ErrLink) formatJSON(format *JSONFormat) (formatted map[string]interface{}) {
	formatted = make(map[string]interface{})

	formatted["message"] = eLink.Message

	if eLink.Type != "" {
		formatted["type"] = eLink.Type
	}

	if len(eLink.Fields) > 0 {
		formatted["fields"] = eLink.Fields
	}

	if format.Options.WithTrace {
		formatted["stack"] = eLink.Frame.format(format.StackElemSep)
	}

	return
}

// FormatOptions controls error rendering behavior across all output formats.
//
// Fields:
//   - InvertOutput (bool): When true, displays wrap chain before root error.
//   - WithTrace (bool): When true, includes stack trace information.
//   - InvertTrace (bool): When true, reverses stack frame order.
//   - WithExternal (bool): When true, includes external errors in output.
type FormatOptions struct {
	InvertOutput bool
	WithTrace    bool
	InvertTrace  bool
	WithExternal bool
}

// StringFormat defines separators and rules for rendering errors as human-readable strings.
// It uses FormatOptions to determine which parts of the error to include.
//
// Fields:
//   - Options (*FormatOptions): controls inclusion of messages, fields, and traces.
//   - TypePrefix (string): prefix for error type in text output.
//   - TypeSuffix (string): suffix for error type.
//   - TypeMessageSeparator (string): separator between type and message.
//   - MessageFieldsSeparator (string): separator before structured fields block.
//   - FieldSep (string): separator between field entries.
//   - FieldKVSep (string): separator between key and value in fields.
//   - FieldsTitleEntriesSeparator (string): separator after "Fields:" title.
//   - FieldsStackSeparator (string): separator before "Stack:" block.
//   - StackTitleEntriesSeparator (string): separator after "Stack:" title.
//   - StackEntriesSeparator (string): separator between individual frames.
//   - PreStackSep (string): prefix for each stack frame line.
//   - StackElemSep (string): separator between elements in a frame line.
//   - ErrorSep  (string): separator between errors in a chain.
type StringFormat struct {
	Options                     *FormatOptions
	TypePrefix                  string
	TypeSuffix                  string
	TypeMessageSeparator        string
	MessageFieldsSeparator      string
	FieldSep                    string
	FieldKVSep                  string
	FieldsTitleEntriesSeparator string
	FieldsStackSeparator        string
	StackTitleEntriesSeparator  string
	StackEntriesSeparator       string
	PreStackSep                 string
	StackElemSep                string
	ErrorSep                    string
}

// JSONFormat defines rules for serializing errors into JSON-like maps.
// It uses FormatOptions to choose which parts to include and how to represent stack elements.
//
// Fields:
//   - Options (*FormatOptions): controls inclusion of external and trace information.
//   - StackElemSep (string): separator between method, file, and line in frame strings.
type JSONFormat struct {
	Options      *FormatOptions
	StackElemSep string
}

// NewDefaultStringFormat returns a StringFormat with default separators based on options.WithTrace.
//
// Parameters:
//   - options (*FormatOptions): flags controlling output sections.
//
// Returns:
//   - format (*StringFormat): the configured format object.
func NewDefaultStringFormat(options *FormatOptions) (format *StringFormat) {
	format = &StringFormat{
		Options: options,
	}

	if options.WithTrace {
		format.TypePrefix = "["
		format.TypeSuffix = "]"
		format.TypeMessageSeparator = " "
		format.MessageFieldsSeparator = "\n\n"
		format.FieldSep = ","
		format.FieldKVSep = "="
		format.FieldsTitleEntriesSeparator = "\n"
		format.FieldsStackSeparator = "\n\n"
		format.StackTitleEntriesSeparator = "\n"
		format.StackEntriesSeparator = "\n"
		format.PreStackSep = "  "
		format.StackElemSep = ":"
		format.ErrorSep = "\n\n"
	} else {
		format.TypePrefix = "["
		format.TypeSuffix = "]"
		format.TypeMessageSeparator = " "
		format.MessageFieldsSeparator = "\n\n"
		format.FieldsTitleEntriesSeparator = "\n"
		format.FieldsStackSeparator = "\n\n"
		format.ErrorSep = "\n\n"
		format.FieldSep = " "
		format.FieldKVSep = "="
	}

	return
}

// ToString formats err into a human-readable string using default StringFormat.
//
// Parameters:
//   - err (error): the error to format.
//   - withTrace (bool): whether to include stack traces.
//
// Returns:
//   - formatted (string): the formatted error string.
func ToString(err error, withTrace bool) (formatted string) {
	return ToCustomString(err, NewDefaultStringFormat(&FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// ToCustomString formats err using the provided StringFormat.
//
// Parameters:
//   - err (error): the error to format.
//   - format (*StringFormat): formatting rules and separators.
//
// Returns:
//   - formatted (string): the formatted error string.
func ToCustomString(err error, format *StringFormat) (formatted string) {
	upErr := Unpack(err)

	if format.Options.InvertOutput {
		if format.Options.WithExternal && upErr.ErrExternal != nil {
			formatted += formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)

			if (format.Options.WithTrace && len(upErr.ErrRoot.Stack) > 0) || upErr.ErrRoot.Message != "" {
				formatted += format.StackEntriesSeparator
			}
		}

		formatted += upErr.ErrRoot.formatStr(format)

		for _, eLink := range upErr.ErrChain {
			formatted += format.ErrorSep + eLink.formatStr(format)
		}
	} else {
		for i := len(upErr.ErrChain) - 1; i >= 0; i-- {
			formatted += upErr.ErrChain[i].formatStr(format) + format.ErrorSep
		}

		formatted += upErr.ErrRoot.formatStr(format)

		if format.Options.WithExternal && upErr.ErrExternal != nil {
			if (format.Options.WithTrace && len(upErr.ErrRoot.Stack) > 0) || upErr.ErrRoot.Message != "" {
				formatted += format.ErrorSep
			}

			formatted += formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
		}
	}

	return
}

// NewDefaultJSONFormat returns a JSONFormat with default separators.
//
// Parameters:
//   - options (*FormatOptions): flags controlling output sections.
//
// Returns:
//   - format (*JSONFormat): the configured format object.
func NewDefaultJSONFormat(options *FormatOptions) (format *JSONFormat) {
	format = &JSONFormat{
		Options:      options,
		StackElemSep: ":",
	}

	return
}

// ToJSON formats err into a JSON-ready map using default JSONFormat.
//
// Parameters:
//   - err (error): the error to format.
//   - withTrace (bool): whether to include stack traces.
//
// Returns:
//   - formatted (map[string]interface{}): the JSON-like representation of the error chain.
func ToJSON(err error, withTrace bool) (formatted map[string]interface{}) {
	return ToCustomJSON(err, NewDefaultJSONFormat(&FormatOptions{
		WithTrace:    withTrace,
		WithExternal: true,
	}))
}

// ToCustomJSON formats err using the provided JSONFormat.
//
// Parameters:
//   - err (error): the error to format.
//   - format (*JSONFormat): formatting rules for JSON serialization.
//
// Returns:
//   - formatted (map[string]interface{}): the JSON-like representation of the error chain.
func ToCustomJSON(err error, format *JSONFormat) (formatted map[string]interface{}) {
	upErr := Unpack(err)

	formatted = make(map[string]interface{})

	if format.Options.WithExternal && upErr.ErrExternal != nil {
		formatted["external"] = formatExternalStr(upErr.ErrExternal, format.Options.WithTrace)
	}

	if upErr.ErrRoot.Message != "" || len(upErr.ErrRoot.Stack) > 0 {
		formatted["root"] = upErr.ErrRoot.formatJSON(format)
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

		formatted["wrap"] = wrapArr
	}

	return
}

// Unpack traverses the error chain by repeatedly calling Unwrap.
// For a rootError, it populates ErrRoot; for each wrapError, it adds an ErrLink;
// any other error becomes ErrExternal.
//
// Parameters:
//   - err (error): the error to decompose.
//
// Returns:
//   - uerr (UnpackedError): a struct containing ErrExternal, ErrRoot, and ErrChain.
func Unpack(err error) (uerr UnpackedError) {
	for err != nil {
		switch err := err.(type) {
		case *rootError:
			uerr.ErrRoot.Message = err.message
			uerr.ErrRoot.Type = err.t
			uerr.ErrRoot.Fields = err.fields
			uerr.ErrRoot.Stack = err.stack.resolveToStackFrames()
		case *wrapError:
			link := ErrLink{
				Message: err.message,
				Type:    err.t,
				Fields:  err.fields,
			}

			link.Frame = err.frame.resolveToStackFrame()

			uerr.ErrChain = append([]ErrLink{link}, uerr.ErrChain...)
		default:
			uerr.ErrExternal = err

			return
		}

		err = Unwrap(err)
	}

	return
}

// formatExternalStr renders an external error to a string.
// If withTrace is true, fmt.Sprintf("%+v", err) is used to include the external error's stack.
//
// Parameters:
//   - err (error): the external error.
//   - withTrace (bool): whether to include the error's own stack trace.
//
// Returns:
//   - s (string): the rendered error string.
func formatExternalStr(err error, withTrace bool) (s string) {
	s = fmt.Sprint(err)

	if withTrace {
		s = fmt.Sprintf("%+v", err)
	}

	return
}
