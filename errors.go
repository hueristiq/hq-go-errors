package errors

import (
	"reflect"
)

// rootError represents a fundamental error with complete stack trace information.
// It serves as the base error type in the package and implements the Error interface.
//
// Fields:
//   - global (bool): indicates if error occurred during package initialization
//   - t (ErrorType): error type for classification (ErrorType)
//   - message (string): human-readable error message
//   - fields (map[string]interface{}): additional structured context (key-value pairs)
//   - wrapped (error): the underlying error being wrapped (if any)
//   - stack (*stack): captured call stack information
type rootError struct {
	global  bool
	t       ErrorType
	message string
	fields  map[string]interface{}
	wrapped error
	stack   *stack
}

// WithType associates a type with the error for classification purposes.
// This enables error handling based on error categories/types.
//
// Parameters:
//   - t (ErrorType): the ErrorType to assign to this error
//
// Returns:
//   - err (Error): the modified error (for method chaining)
func (e *rootError) WithType(t ErrorType) (err Error) {
	e.t = t

	err = e

	return
}

// Type returns the error's classification type if one was set.
//
// Returns:
//   - t (ErrorType): the error's type, or empty string if untyped
func (e rootError) Type() (t ErrorType) {
	t = e.t

	return
}

// WithField adds a key-value pair to the error's structured context.
// Fields provide additional machine-readable information about the error.
//
// Parameters:
//   - key (string): field name (should be descriptive and consistent)
//   - value (interface{}): field value (any serializable type)
//
// Returns:
//   - err (Error): the modified error (for method chaining)
func (e *rootError) WithField(key string, value interface{}) (err Error) {
	if e.fields == nil {
		e.fields = make(map[string]interface{})
	}

	e.fields[key] = value

	err = e

	return
}

// Fields returns all structured fields attached to the error.
// The returned map should not be modified directly.
//
// Returns:
//   - fields (map[string]interface{}): all attached fields (may be nil)
func (e *rootError) Fields() (fields map[string]interface{}) {
	fields = e.fields

	return
}

// Error implements the error interface, returning the error message.
// If the error wraps another error, it combines both messages.
//
// Returns:
//   - s (string): the error message (or "<nil>" if receiver is nil)
func (e *rootError) Error() (s string) {
	s = "<nil>"

	if e == nil {
		return
	}

	s = e.message

	if e.wrapped != nil {
		s += ": " + e.wrapped.Error()
	}

	return
}

// Is implements error equality checking. Two errors are considered equal if:
//   - Both are nil, or
//   - They are of the same type (*rootError), and:
//   - Their types match (or target type is empty), and
//   - Their messages match
//   - Their messages match exactly (fallback)
//
// Parameters:
//   - target (error): the error to compare against
//
// Returns:
//   - is (bool): true if errors are considered equal
func (e *rootError) Is(target error) (is bool) {
	if target == nil {
		is = e == nil

		return
	}

	if err, ok := target.(*rootError); ok {
		is = (err.t == "" || e.t == err.t) && e.message == err.message

		return
	}

	is = e.message == target.Error()

	return
}

// As attempts to assign the error to the target interface.
// The target must be a non-nil pointer to either:
//   - An interface type that the error implements, or
//   - A concrete type that matches the error's type
//
// Parameters:
//   - target (interface{}): pointer to interface or concrete type
//
// Returns:
//   - as (bool): true if assignment was successful
func (e *rootError) As(target interface{}) (as bool) {
	if target == nil {
		return
	}

	val := reflect.ValueOf(target)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}

	targetType := val.Type().Elem()
	currentType := reflect.TypeOf(e)

	if currentType.AssignableTo(targetType) {
		val.Elem().Set(reflect.ValueOf(e))

		as = true

		return
	}

	return
}

// Unwrap returns the underlying error if this error wraps another.
// Implements the standard library's error unwrapping interface.
//
// Returns:
//   - err (error): the wrapped error (may be nil)
func (e rootError) Unwrap() (err error) {
	err = e.wrapped

	return
}

// StackFrames returns the raw PCs (program counters) from the call stack.
// These can be used to reconstruct the full stack trace.
//
// Returns:
//   - PCs ([]uintptr): slice of program counters representing the call stack
func (e *rootError) StackFrames() (PCs []uintptr) {
	PCs = *e.stack

	return
}

// wrapError represents an error that wraps another error with additional context.
// Unlike rootError, it only captures a single stack frame (where it was created).
//
// Fields:
//   - t (ErrorType): error type for classification (ErrorType)
//   - message (string): human-readable error message
//   - fields (map[string]interface{}): additional structured context (key-value pairs)
//   - err (error): underlying error being wrapped
//   - frame (*frame): stack frame where the wrap occurred
type wrapError struct {
	t       ErrorType
	message string
	fields  map[string]interface{}
	err     error
	frame   *frame
}

// WithType associates a type with the error for classification purposes.
// This enables error handling based on error categories/types.
//
// Parameters:
//   - t (ErrorType): the ErrorType to assign to this error
//
// Returns:
//   - err (Error): the modified error (for method chaining)
func (e *wrapError) WithType(t ErrorType) (err Error) {
	e.t = t

	err = e

	return
}

// Type returns the error's classification type if one was set.
//
// Returns:
//   - t (ErrorType): the error's type, or empty string if untyped
func (e wrapError) Type() (t ErrorType) {
	t = e.t

	return
}

// WithField adds a key-value pair to the error's structured context.
// Fields provide additional machine-readable information about the error.
//
// Parameters:
//   - key (string): field name (should be descriptive and consistent)
//   - value (interface{}): field value (any serializable type)
//
// Returns:
//   - (Error): the modified error (for method chaining)
func (e *wrapError) WithField(key string, value interface{}) (err Error) {
	if e.fields == nil {
		e.fields = make(map[string]interface{})
	}

	e.fields[key] = value

	err = e

	return
}

// Fields returns all structured fields attached to the error.
// The returned map should not be modified directly.
//
// Returns:
//   - fields (map[string]interface{}): all attached fields (may be nil)
func (e *wrapError) Fields() (fields map[string]interface{}) {
	fields = e.fields

	return
}

// Error implements the error interface, returning the error message.
// If the error wraps another error, it combines both messages.
//
// Returns:
//   - s (string): the error message (or "<nil>" if receiver is nil)
func (e *wrapError) Error() (s string) {
	s = "<nil>"

	if e == nil {
		return
	}

	s = e.message

	if e.err != nil {
		s += ": " + e.err.Error()
	}

	return
}

// Is implements error equality checking. Two errors are considered equal if:
//   - Both are nil, or
//   - They are of the same type (*wrapError), and:
//   - Their types match (or target type is empty), and
//   - Their messages match
//   - Their messages match exactly (fallback)
//
// Parameters:
//   - target (error): the error to compare against
//
// Returns:
//   - is (bool): true if errors are considered equal
func (e *wrapError) Is(target error) (is bool) {
	if target == nil {
		is = e == nil

		return
	}

	if err, ok := target.(*wrapError); ok {
		is = (err.t == "" || e.t == err.t) && e.message == err.message

		return
	}

	is = e.message == target.Error()

	return
}

// As attempts to assign the error to the target interface.
// The target must be a non-nil pointer to either:
//   - An interface type that the error implements, or
//   - A concrete type that matches the error's type
//
// Parameters:
//   - target (interface{}): pointer to interface or concrete type
//
// Returns:
//   - as (bool): true if assignment was successful
func (e *wrapError) As(target interface{}) (as bool) {
	if target == nil {
		return
	}

	val := reflect.ValueOf(target)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}

	targetType := val.Type().Elem()
	currentType := reflect.TypeOf(e)

	if currentType.AssignableTo(targetType) {
		val.Elem().Set(reflect.ValueOf(e))

		as = true

		return
	}

	return
}

// Unwrap returns the underlying error if this error wraps another.
// Implements the standard library's error unwrapping interface.
//
// Returns:
//   - (error): the wrapped error (may be nil)
func (e wrapError) Unwrap() (err error) {
	err = e.err

	return
}

// StackFrames returns the raw program counters from the call stack.
// These can be used to reconstruct the full stack trace.
//
// Returns:
//   - PCs ([]uintptr): slice of program counters representing the call stack
func (e *wrapError) StackFrames() (PCs []uintptr) {
	PCs = []uintptr{e.frame.pc()}

	return
}

type Error interface {
	error
	WithType(ErrorType) Error
	Type() ErrorType
	WithField(string, interface{}) Error
	Fields() map[string]interface{}
	Is(error) bool
	As(interface{}) bool
	Unwrap() error
	StackFrames() []uintptr
}

// ErrorType represents a classification type for errors.
// Types allow errors to be categorized and handled based on their kind.
type ErrorType string

// ErrorOption represents a function that can configure an Error.
// Used with New and Wrap to set error properties at creation time.
type ErrorOption func(Error)

var (
	_ Error = (*rootError)(nil)
	_ Error = (*wrapError)(nil)
)

// New creates a new rootError error with stack trace information.
// The skip parameter (3) ensures the trace starts at the caller's location.
//
// Parameters:
//   - message (string): the primary error message
//   - options (...ErrorOption): variadic list of ErrorOption functions to configure the error
//
// Returns:
//   - err (error): the newly created error (implements Error interface)
func New(message string, options ...ErrorOption) (err error) {
	stack := callers(3)

	e := &rootError{
		global:  stack.isGlobal(),
		message: message,
		stack:   stack,
	}

	for _, option := range options {
		option(e)
	}

	err = e

	return
}

// WithType creates an ErrorOption that sets an error's type.
//
// Parameters:
//   - t (ErrorType): the ErrorType to set
//
// Returns:
//   - option (ErrorOption): configuration function for New/Wrap
func WithType(t ErrorType) (option ErrorOption) {
	return func(e Error) {
		e.WithType(t)
	}
}

// WithField creates an ErrorOption that adds a field to an error.
//
// Parameters:
//   - k (string): field key
//   - v (interface{}): field value
//
// Returns:
//   - option ErrorOption: configuration function for New/Wrap
func WithField(k string, v interface{}) (option ErrorOption) {
	return func(e Error) {
		e.WithField(k, v)
	}
}

// Wrap creates a new error that wraps an existing error with additional context.
// The new error will have its own stack frame while preserving the original's trace.
//
// Parameters:
//   - err (error): the error to wrap
//   - message (string): additional context message
//   - options (...ErrorOption): configuration options (same as New)
//
// Returns:
//   - errr (error): the new wrapping error
func Wrap(err error, message string, options ...ErrorOption) (errr error) {
	w := wrap(err, message)

	for _, option := range options {
		option(w)
	}

	errr = w

	return
}

// wrap is the internal implementation of error wrapping logic.
// Handles three cases:
//  1. Wrapping a root error (preserves full stack)
//  2. Wrapping a wrapped error (finds root to preserve stack)
//  3. Wrapping a non-package error (creates new root)
func wrap(err error, message string) (errr Error) {
	if err == nil {
		return nil
	}

	stack := callers(4)
	frame := caller(3)

	switch e := err.(type) {
	case *rootError:
		if e.global {
			err = &rootError{
				global:  e.global,
				message: e.message,
				stack:   stack,
			}
		} else {
			e.stack.insertPC(*stack)
		}
	case *wrapError:
		if r, ok := Cause(err).(*rootError); ok {
			r.stack.insertPC(*stack)
		}
	default:
		errr = &rootError{
			message: message,
			wrapped: e,
			stack:   stack,
		}

		return
	}

	errr = &wrapError{
		message: message,
		err:     err,
		frame:   frame,
	}

	return
}

// Unwrap returns the result of calling Unwrap() on err if available.
// Matches the behavior of errors.Unwrap in the standard library.
//
// Parameters:
//   - err (error): the error to unwrap.
//
// Returns:
//   - errr (error): The next error in the chain, or nil if none.
func Unwrap(err error) (errr error) {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return
	}

	return u.Unwrap()
}

// Is reports whether err or any error in its chain matches target.
// Implements an enhanced version of errors.Is from the standard library.
//
// Parameters:
//   - err (error): the error to inspect.
//   - target (error): the error to compare against.
//
// Returns:
//   - is (bool): true if any error in err's chain matches target.
func Is(err, target error) (is bool) {
	if target == nil {
		is = err == target

		return
	}

	isComparable := reflect.TypeOf(target).Comparable()

	for {
		if isComparable && err == target {
			is = true

			return
		}

		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			is = true

			return
		}

		if x, ok := err.(interface{ Is(error) bool }); ok {
			if x.Is(target) {
				is = true

				return
			}
		}

		if err = Unwrap(err); err == nil {
			return
		}
	}
}

// As searches err's chain for an error assignable to target and sets target if found.
// Implements an enhanced version of errors.As from the standard library.
//
// Parameters:
//   - err (error): the error to inspect.
//   - target (interface{}): pointer to the destination interface or concrete type.
//
// Returns:
//   - as (bool): true if a matching error was found and target was set.
func As(err error, target interface{}) (as bool) {
	if target == nil || err == nil {
		return
	}

	val := reflect.ValueOf(target)
	typ := val.Type()

	if typ.Kind() != reflect.Ptr || val.IsNil() {
		return
	}

	targetType := typ.Elem()

	if targetType.Kind() != reflect.Interface && !targetType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return
	}

	for {
		// Try custom As method first
		if x, ok := err.(interface{ As(interface{}) bool }); ok {
			if x.As(target) {
				as = true

				return
			}
		}

		// Standard type assignment check
		if reflect.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(err))

			as = true

			return
		}

		if err = Unwrap(err); err == nil {
			return
		}
	}
}

// Cause returns the underlying root cause of the error by recursively unwrapping.
// Unlike Unwrap, it follows the entire chain to the original error.
//
// Parameters:
//   - err (error): the error to inspect.
//
// Returns:
//   - errr (error): The deepest non-wrapped error in the chain.
func Cause(err error) (errr error) {
	for {
		uerr := Unwrap(err)
		if uerr == nil {
			return err
		}

		err = uerr
	}
}
