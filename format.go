package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UnpackedError represents the decomposed structure of an error.
// It breaks down complex errors into their constituent parts for easier formatting and analysis.
// This struct is used internally by formatting functions to organize error information.
//
// Fields:
//   - ErrExternal (error): any external (non-package) error found in the chain
//   - ErrRoot (ErrPart): the root error part, if present
//   - ErrChain ([]ErrPart): the chain of wrapped error parts
//   - ErrJoined ([]error): list of joined errors, if the error is a joined type
type UnpackedError struct {
	ErrExternal error
	ErrRoot     ErrPart
	ErrChain    []ErrPart
	ErrJoined   []error
}

// ErrPart represents a single component of an error, either root or wrapped.
// It encapsulates the key details of that error segment for formatting.
//
// Fields:
//   - Message (string): the error message for this part
//   - Type (Type): the classification type of this error part
//   - Fields (map[string]any): structured key-value fields associated with this part
//   - Stack (Stack): the stack trace frames for this error part
type ErrPart struct {
	Message string
	Type    Type
	Fields  map[string]any
	Stack   Stack
}

// Formatter is responsible for converting errors into human-readable string or JSON formats.
// It uses configurable options to control the output structure and content.
//
// Fields:
//   - options (*FormatterOptions): the configuration options for formatting
type Formatter struct {
	options *FormatterOptions
}

// String formats the error as a multi-line string.
// It handles both chain and joined errors differently.
//
// Parameters:
//   - err (error): the error to format
//
// Returns:
//   - formated (string): the formatted string representation, or empty if err is nil
func (f *Formatter) String(err error) (formated string) {
	if err == nil {
		return
	}

	switch e := err.(type) {
	case *joined:
		formated = f.formatJoinedString(e)

		return
	default:
		formated = f.formatChainString(err)

		return
	}
}

// JSON formats the error as a map suitable for JSON encoding.
// It handles both chain and joined errors differently.
//
// Parameters:
//   - err (error): the error to format
//
// Returns:
//   - formated (map[string]any): the formatted map, or nil if err is nil
func (f *Formatter) JSON(err error) (formated map[string]any) {
	if err == nil {
		return
	}

	switch e := err.(type) {
	case *joined:
		formated = f.formatJoinedJSON(e)

		return
	default:
		formated = f.formatChainJSON(err)

		return
	}
}

// formatChainString formats a chain error (root + wraps) into a string.
// It unpacks the error and assembles parts based on options (e.g., order, external inclusion).
//
// Parameters:
//   - err (error): the chain error to format
//
// Returns:
//   - (string): the formatted string
func (f *Formatter) formatChainString(err error) string {
	unpacked := Unpack(err)

	var parts []string

	if f.options.IsInnerFirst {
		if unpacked.ErrExternal != nil && (f.options.WithExternal || f.isOnlyExternal(&unpacked)) {
			parts = append(parts, f.formatExternalString(unpacked.ErrExternal))
		}

		if f.hasRootContent(&unpacked.ErrRoot) {
			parts = append(parts, f.formatPartString(&unpacked.ErrRoot, "root"))
		}

		for i := len(unpacked.ErrChain) - 1; i >= 0; i-- {
			parts = append(parts, f.formatPartString(&unpacked.ErrChain[i], "wrap"))
		}
	} else {
		for i := range len(unpacked.ErrChain) {
			parts = append(parts, f.formatPartString(&unpacked.ErrChain[i], "wrap"))
		}

		if f.hasRootContent(&unpacked.ErrRoot) {
			parts = append(parts, f.formatPartString(&unpacked.ErrRoot, "root"))
		}

		if unpacked.ErrExternal != nil && (f.options.WithExternal || f.isOnlyExternal(&unpacked)) {
			parts = append(parts, f.formatExternalString(unpacked.ErrExternal))
		}
	}

	separator := "\n\n"

	return strings.Join(parts, separator)
}

// formatPartString formats a single ErrPart into a string.
// It includes type, message, fields, and optional trace.
//
// Parameters:
//   - part (*ErrPart): the error part to format
//   - kind (string): the kind of part ("root" or "wrap") for trace labeling
//
// Returns:
//   - (string): the formatted string for this part
func (f *Formatter) formatPartString(part *ErrPart, kind string) string {
	var buf strings.Builder

	if part.Type != "" {
		buf.WriteString("[")
		buf.WriteString(string(part.Type))
		buf.WriteString("]" + f.options.Spacing)
	}

	buf.WriteString(part.Message)

	if len(part.Fields) > 0 {
		buf.WriteString("\n\nFields:")

		for k, v := range part.Fields {
			buf.WriteString(fmt.Sprintf("\n%s%s:%s%v", f.options.Indentation, k, f.options.Spacing, v))
		}
	}

	if f.options.WithTrace && len(part.Stack) > 0 {
		frames := part.Stack

		buf.WriteString(fmt.Sprintf("\n\n%s Trace:", kind))

		for _, frame := range frames {
			buf.WriteString(fmt.Sprintf("\n%s%s%s(%s:%d)", f.options.Indentation, frame.Name, f.options.Spacing, frame.File, frame.Line))
		}
	}

	return buf.String()
}

// formatExternalString formats an external error into a string.
// It includes trace if configured, otherwise just the error message.
//
// Parameters:
//   - err (error): the external error to format
//
// Returns:
//   - (string): the formatted string
func (f *Formatter) formatExternalString(err error) string {
	if f.options.WithTrace {
		return fmt.Sprintf("%+v", err)
	}

	return err.Error()
}

// formatJoinedString formats a joined error into a string.
// It includes the count, optional join location, and formats each sub-error recursively.
//
// Parameters:
//   - joinErr (*joined): the joined error to format
//
// Returns:
//   - (string): the formatted string
func (f *Formatter) formatJoinedString(joinErr *joined) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("Multiple errors (%d):", len(joinErr.errors)))

	if f.options.WithTrace && joinErr.trace != nil {
		frames := joinErr.trace.resolveToStackFrames()

		if len(frames) > 0 {
			buf.WriteString("\n\nJoin Location:")

			if len(frames) > 0 {
				frame := frames[0]

				buf.WriteString(fmt.Sprintf("\n%s%s%s(%s:%d)", f.options.Indentation, frame.Name, f.options.Spacing, frame.File, frame.Line))
			}
		}
	}

	for i, err := range joinErr.errors {
		if err == nil {
			continue
		}

		buf.WriteString(fmt.Sprintf("\n\n%d. %s", i+1, f.String(err)))
	}

	return buf.String()
}

// formatChainJSON formats a chain error into a JSON-compatible map.
// It unpacks the error and structures it with optional reversal based on options.
//
// Parameters:
//   - err (error): the chain error to format
//
// Returns:
//   - (map[string]any): the formatted map
func (f *Formatter) formatChainJSON(err error) map[string]any {
	unpacked := Unpack(err)
	result := make(map[string]any)

	if unpacked.ErrExternal != nil && (f.options.WithExternal || f.isOnlyExternal(&unpacked)) {
		result["external"] = map[string]any{
			"message": unpacked.ErrExternal.Error(),
			"go_type": fmt.Sprintf("%T", unpacked.ErrExternal),
		}
	}

	if f.hasRootContent(&unpacked.ErrRoot) {
		result["root"] = f.formatPartJSON(&unpacked.ErrRoot)
	}

	if len(unpacked.ErrChain) > 0 {
		var chain []map[string]any

		for _, part := range unpacked.ErrChain {
			chain = append(chain, f.formatPartJSON(&part))
		}

		if f.options.IsInnerFirst {
			for i := len(chain)/2 - 1; i >= 0; i-- {
				opp := len(chain) - 1 - i

				chain[i], chain[opp] = chain[opp], chain[i]
			}
		}

		result["chain"] = chain
	}

	return result
}

// formatPartJSON formats a single ErrPart into a JSON-compatible map.
// It includes message, type, fields, and optional stack with possible inversion.
//
// Parameters:
//   - part (*ErrPart): the error part to format
//
// Returns:
//   - (map[string]any): the formatted map
func (f *Formatter) formatPartJSON(part *ErrPart) map[string]any {
	result := map[string]any{
		"message": part.Message,
	}

	if part.Type != "" {
		result["type"] = string(part.Type)
	}

	if len(part.Fields) > 0 {
		result["fields"] = part.Fields
	}

	if f.options.WithTrace && len(part.Stack) > 0 {
		var frames []map[string]any

		stack := part.Stack

		for _, frame := range stack {
			frameMap := map[string]any{
				"function": frame.Name,
				"file":     frame.File,
				"line":     frame.Line,
			}

			frames = append(frames, frameMap)
		}

		if f.options.InvertTrace {
			for i := len(frames)/2 - 1; i >= 0; i-- {
				opp := len(frames) - 1 - i

				frames[i], frames[opp] = frames[opp], frames[i]
			}
		}

		result["stack"] = frames
	}

	return result
}

// formatJoinedJSON formats a joined error into a JSON-compatible map.
// It includes type, count, optional join stack, and recursively formats sub-errors.
//
// Parameters:
//   - joinErr (*joined): the joined error to format
//
// Returns:
//   - (map[string]any): the formatted map
func (f *Formatter) formatJoinedJSON(joinErr *joined) map[string]any {
	result := map[string]any{
		"type":  "joined",
		"count": len(joinErr.errors),
	}

	if f.options.WithTrace && joinErr.trace != nil {
		frames := joinErr.trace.resolveToStackFrames()

		if len(frames) > 0 {
			var joinFrames []map[string]any

			for _, frame := range frames {
				joinFrames = append(joinFrames, map[string]any{
					"function": frame.Name,
					"file":     frame.File,
					"line":     frame.Line,
				})
			}

			result["join_stack"] = joinFrames
		}
	}

	var errors []any

	for _, err := range joinErr.errors {
		if err != nil {
			errors = append(errors, f.JSON(err))
		}
	}

	result["errors"] = errors

	return result
}

// hasRootContent checks if the root ErrPart has any meaningful content.
// Used to decide whether to include the root in formatting.
//
// Parameters:
//   - root (*ErrPart): the root part to check
//
// Returns:
//   - (bool): true if it has a message or stack frames
func (f *Formatter) hasRootContent(root *ErrPart) bool {
	return root.Message != "" || len(root.Stack) > 0
}

// isOnlyExternal checks if the unpacked error consists only of an external error.
// Used to decide inclusion when WithExternal is false.
//
// Parameters:
//   - unpacked (*UnpackedError): the unpacked error to check
//
// Returns:
//   - (bool): true if only external error is present
func (f *Formatter) isOnlyExternal(unpacked *UnpackedError) bool {
	return unpacked.ErrExternal != nil && unpacked.ErrRoot.Message == "" && len(unpacked.ErrChain) == 0
}

// FormatterOptions holds configuration for the Formatter.
// It controls aspects like order, trace inclusion, and formatting style.
//
// Fields:
//   - IsInnerFirst (bool): if true, format from inner to outer (default: false)
//   - WithTrace (bool): include stack traces (default: false)
//   - InvertTrace (bool): invert stack trace order (default: false)
//   - WithExternal (bool): include external errors (default: true)
//   - Spacing (string): spacing between elements (default: " ")
//   - Indentation (string): indentation for nested elements (default: "  ")
type FormatterOptions struct {
	IsInnerFirst bool
	WithTrace    bool
	InvertTrace  bool
	WithExternal bool
	Spacing      string
	Indentation  string
}

// FormatterOptionFunc is a function type for configuring FormatterOptions.
// Used with NewFormatter to set custom options.
type FormatterOptionFunc func(options *FormatterOptions)

// NewFormatter creates a new Formatter with default or custom options.
// Defaults: outer-first, no trace, no invert, include external, space " ", indent "  ".
//
// Parameters:
//   - ofs (...FormatterOptionFunc): variadic option functions
//
// Returns:
//   - formatter (*Formatter): the new formatter instance
func NewFormatter(ofs ...FormatterOptionFunc) (formatter *Formatter) {
	options := &FormatterOptions{
		IsInnerFirst: false,
		WithTrace:    false,
		InvertTrace:  false,
		WithExternal: true,
		Spacing:      " ",
		Indentation:  "  ",
	}

	for _, f := range ofs {
		f(options)
	}

	formatter = &Formatter{
		options: options,
	}

	return
}

// FormatWithTrace returns an option function to enable stack traces.
func FormatWithTrace() (f FormatterOptionFunc) {
	return func(options *FormatterOptions) {
		options.WithTrace = true
	}
}

// Unpack decomposes an error into its parts.
// It handles joined, root, wrapped, and external errors.
//
// The unpacking process:
//  1. If joined, sets ErrJoined and returns.
//  2. Traverses the chain using Unwrap.
//  3. For root/wrapped, extracts to ErrRoot/ErrChain.
//  4. For external, sets ErrExternal.
//
// Parameters:
//   - err (error): the error to unpack
//
// Returns:
//   - uerr (UnpackedError): the unpacked structure
func Unpack(err error) (uerr UnpackedError) {
	if joinErr, ok := err.(*joined); ok {
		uerr.ErrJoined = joinErr.errors

		return
	}

	for err != nil {
		switch e := err.(type) {
		case *root:
			uerr.ErrRoot = ErrPart{
				Type:    e.errType,
				Message: e.message,
				Fields:  e.fields,
			}

			if e.trace != nil {
				uerr.ErrRoot.Stack = e.trace.resolveToStackFrames()
			}
		case *wrapped:
			part := ErrPart{
				Type:    e.errType,
				Message: e.message,
				Fields:  e.fields,
			}

			if e.frame != nil {
				part.Stack = Stack{e.frame.resolveToStackFrame()}
			}

			uerr.ErrChain = append(uerr.ErrChain, part)
		default:
			uerr.ErrExternal = err

			return
		}

		err = Unwrap(err)
	}

	return
}

// ToString is a convenience function to format an error as a string.
// It creates a formatter with options and calls String.
//
// Parameters:
//   - err (error): the error to format
//   - ofs (...FormatterOptionFunc): optional configuration
//
// Returns:
//   - formated (string): the formatted string
func ToString(err error, ofs ...FormatterOptionFunc) (formated string) {
	formatter := NewFormatter(ofs...)

	formated = formatter.String(err)

	return
}

// ToJSON is a convenience function to format an error as a JSON map.
// It creates a formatter with options and calls JSON.
//
// Parameters:
//   - err (error): the error to format
//   - ofs (...FormatterOptionFunc): optional configuration
//
// Returns:
//   - formated (map[string]any): the formatted map
func ToJSON(err error, ofs ...FormatterOptionFunc) (formated map[string]any) {
	formatter := NewFormatter(ofs...)

	formated = formatter.JSON(err)

	return
}

// ToJSONString is a convenience function to format an error as a JSON string.
// It uses ToJSON and marshals with indentation.
//
// Parameters:
//   - err (error): the error to format
//   - ofs (...FormatterOptionFunc): optional configuration
//
// Returns:
//   - formated (string): the JSON string, or error message if marshaling fails
func ToJSONString(err error, ofs ...FormatterOptionFunc) (formated string) {
	data := ToJSON(err, ofs...)
	if data == nil {
		return
	}

	bytes, jsonErr := json.MarshalIndent(data, "", "  ")
	if jsonErr != nil {
		formated = fmt.Sprintf("JSON formatting error: %v", jsonErr)

		return
	}

	formated = string(bytes)

	return
}
