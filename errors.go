package errors

import (
	"reflect"
)

// root represents a fundamental error with complete stack trace information.
// It serves as the base error type in the package and implements the Error interface.
//
// Fields:
//   - isDeclaredGlobally bool: indicates if error occurred during package initialization
//   - errorType ErrorType: error type for classification (ErrorType)
//   - message string: human-readable error message
//   - fields map[string]interface{}: additional structured context (key-value pairs)
//   - wrapped error: the underlying error being wrapped (if any)
//   - stack *stack: captured call stack information
type root struct {
	isDeclaredGlobally bool

	errorType ErrorType
	message   string
	fields    map[string]interface{}

	wrapped error

	stack *stack
}

// WithType associates a type with the error for classification purposes.
// This enables error handling based on error categories/types.
//
// Parameters:
//   - errorType ErrorType: the ErrorType to assign to this error
//
// Returns:
//   - Error: the modified error (for method chaining)
func (e *root) WithType(errorType ErrorType) Error {
	e.errorType = errorType

	return e
}

// Type returns the error's classification type if one was set.
//
// Returns:
//   - ErrorType: the error's type, or empty string if untyped
func (e root) Type() ErrorType {
	return e.errorType
}

// WithField adds a key-value pair to the error's structured context.
// Fields provide additional machine-readable information about the error.
//
// Parameters:
//   - key string: field name (should be descriptive and consistent)
//   - value interface{}: field value (any serializable type)
//
// Returns:
//   - Error: the modified error (for method chaining)
func (e *root) WithField(key string, value interface{}) Error {
	if e.fields == nil {
		e.fields = make(map[string]interface{})
	}

	e.fields[key] = value

	return e
}

// Fields returns all structured fields attached to the error.
// The returned map should not be modified directly.
//
// Returns:
//   - map[string]interface{}: all attached fields (may be nil)
func (e *root) Fields() map[string]interface{} {
	return e.fields
}

// Error implements the error interface, returning the error message.
// If the error wraps another error, it combines both messages.
//
// Returns:
//   - string: the error message (or "<nil>" if receiver is nil)
func (e *root) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.wrapped != nil {
		return e.message + ": " + e.wrapped.Error()
	}

	return e.message
}

// Is implements error equality checking. Two errors are considered equal if:
//   - Both are nil, or
//   - They are of the same type (*root), and:
//   - Their types match (or target type is empty), and
//   - Their messages match
//   - Their messages match exactly (fallback)
//
// Parameters:
//   - target error: the error to compare against
//
// Returns:
//   - bool: true if errors are considered equal
func (e *root) Is(target error) bool {
	if target == nil {
		return e == nil
	}

	if err, ok := target.(*root); ok {
		return (err.errorType == "" || e.errorType == err.errorType) && e.message == err.message
	}

	return e.message == target.Error()
}

// As attempts to assign the error to the target interface.
// The target must be a non-nil pointer to either:
//   - An interface type that the error implements, or
//   - A concrete type that matches the error's type
//
// Parameters:
//   - target interface{}: pointer to interface or concrete type
//
// Returns:
//   - bool: true if assignment was successful
func (e *root) As(target interface{}) bool {
	if target == nil {
		return false
	}

	val := reflect.ValueOf(target)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	targetType := val.Type().Elem()
	currentType := reflect.TypeOf(e)

	if currentType.AssignableTo(targetType) {
		val.Elem().Set(reflect.ValueOf(e))

		return true
	}

	return false
}

// Unwrap returns the underlying error if this error wraps another.
// Implements the standard library's error unwrapping interface.
//
// Returns:
//   - error: the wrapped error (may be nil)
func (e root) Unwrap() error {
	return e.wrapped
}

// StackFrames returns the raw program counters from the call stack.
// These can be used to reconstruct the full stack trace.
//
// Returns:
//   - []uintptr: slice of program counters representing the call stack
func (e *root) StackFrames() []uintptr {
	return *e.stack
}

// wrapped represents an error that wraps another error with additional context.
// Unlike root, it only captures a single stack frame (where it was created).
//
// Fields:
//   - errorType ErrorType: error type for classification (ErrorType)
//   - message string: human-readable error message
//   - fields map[string]interface{}: additional structured context (key-value pairs)
//   - err error: underlying error being wrapped
//   - frame *frame: stack frame where the wrap occurred
type wrapped struct {
	errorType ErrorType
	message   string
	fields    map[string]interface{}

	err error

	frame *frame
}

// WithType associates a type with the error for classification purposes.
// This enables error handling based on error categories/types.
//
// Parameters:
//   - errorType ErrorType: the ErrorType to assign to this error
//
// Returns:
//   - Error: the modified error (for method chaining)
func (e *wrapped) WithType(errorType ErrorType) Error {
	e.errorType = errorType

	return e
}

// Type returns the error's classification type if one was set.
//
// Returns:
//   - ErrorType: the error's type, or empty string if untyped
func (e wrapped) Type() ErrorType {
	return e.errorType
}

// WithField adds a key-value pair to the error's structured context.
// Fields provide additional machine-readable information about the error.
//
// Parameters:
//   - key string: field name (should be descriptive and consistent)
//   - value interface{}: field value (any serializable type)
//
// Returns:
//   - Error: the modified error (for method chaining)
func (e *wrapped) WithField(key string, value interface{}) Error {
	if e.fields == nil {
		e.fields = make(map[string]interface{})
	}

	e.fields[key] = value

	return e
}

// Fields returns all structured fields attached to the error.
// The returned map should not be modified directly.
//
// Returns:
//   - map[string]interface{}: all attached fields (may be nil)
func (e *wrapped) Fields() map[string]interface{} {
	return e.fields
}

// Error implements the error interface, returning the error message.
// If the error wraps another error, it combines both messages.
//
// Returns:
//   - string: the error message (or "<nil>" if receiver is nil)
func (e *wrapped) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.err != nil {
		return e.message + ": " + e.err.Error()
	}

	return e.message
}

// Is implements error equality checking. Two errors are considered equal if:
//   - Both are nil, or
//   - They are of the same type (*wrapped), and:
//   - Their types match (or target type is empty), and
//   - Their messages match
//   - Their messages match exactly (fallback)
//
// Parameters:
//   - target error: the error to compare against
//
// Returns:
//   - bool: true if errors are considered equal
func (e *wrapped) Is(target error) bool {
	if target == nil {
		return e == nil
	}

	if err, ok := target.(*wrapped); ok {
		return (err.errorType == "" || e.errorType == err.errorType) && e.message == err.message
	}

	return e.message == target.Error()
}

// As attempts to assign the error to the target interface.
// The target must be a non-nil pointer to either:
//   - An interface type that the error implements, or
//   - A concrete type that matches the error's type
//
// Parameters:
//   - target interface{}: pointer to interface or concrete type
//
// Returns:
//   - bool: true if assignment was successful
func (e *wrapped) As(target interface{}) bool {
	if target == nil {
		return false
	}

	val := reflect.ValueOf(target)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	targetType := val.Type().Elem()
	currentType := reflect.TypeOf(e)

	if currentType.AssignableTo(targetType) {
		val.Elem().Set(reflect.ValueOf(e))

		return true
	}

	return false
}

// Unwrap returns the underlying error if this error wraps another.
// Implements the standard library's error unwrapping interface.
//
// Returns:
//   - error: the wrapped error (may be nil)
func (e wrapped) Unwrap() error {
	return e.err
}

// StackFrames returns the raw program counters from the call stack.
// These can be used to reconstruct the full stack trace.
//
// Returns:
//   - []uintptr: slice of program counters representing the call stack
func (e *wrapped) StackFrames() []uintptr {
	return []uintptr{e.frame.pc()}
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

// Option represents a function that can configure an Error.
// Used with New and Wrap to set error properties at creation time.
type Option func(Error)

var (
	_ Error = (*root)(nil)
	_ Error = (*wrapped)(nil)
)

// New creates a new root error with stack trace information.
// The skip parameter (3) ensures the trace starts at the caller's location.
//
// Parameters:
//   - message string: the primary error message
//   - options ...Option: variadic list of Option functions to configure the error
//
// Returns:
//   - error: the newly created error (implements Error interface)
func New(message string, options ...Option) error {
	stack := callers(3)

	e := &root{
		isDeclaredGlobally: stack.isGlobal(),
		message:            message,
		stack:              stack,
	}

	for _, option := range options {
		option(e)
	}

	return e
}

// WithType creates an Option that sets an error's type.
//
// Parameters:
//   - t ErrorType: the ErrorType to set
//
// Returns:
//   - option Option: configuration function for New/Wrap
func WithType(t ErrorType) (option Option) {
	return func(e Error) {
		e.WithType(t)
	}
}

// WithField creates an Option that adds a field to an error.
//
// Parameters:
//   - k string: field key
//   - v interface{}: field value
//
// Returns:
//   - option Option: configuration function for New/Wrap
func WithField(k string, v interface{}) (option Option) {
	return func(e Error) {
		e.WithField(k, v)
	}
}

// Wrap creates a new error that wraps an existing error with additional context.
// The new error will have its own stack frame while preserving the original's trace.
//
// Parameters:
//   - err error: the error to wrap
//   - message string: additional context message
//   - options ...Option: configuration options (same as New)
//
// Returns:
//   - error: the new wrapping error
func Wrap(err error, message string, options ...Option) error {
	w := wrap(err, message)

	for _, option := range options {
		option(w)
	}

	return w
}

// wrap is the internal implementation of error wrapping logic.
// Handles three cases:
//  1. Wrapping a root error (preserves full stack)
//  2. Wrapping a wrapped error (finds root to preserve stack)
//  3. Wrapping a non-package error (creates new root)
func wrap(err error, message string) Error {
	if err == nil {
		return nil
	}

	stack := callers(4)
	frame := caller(3)

	switch e := err.(type) {
	case *root:
		if e.isDeclaredGlobally {
			err = &root{
				isDeclaredGlobally: e.isDeclaredGlobally,
				message:            e.message,
				stack:              stack,
			}
		} else {
			e.stack.insertPC(*stack)
		}
	case *wrapped:
		if r, ok := Cause(err).(*root); ok {
			r.stack.insertPC(*stack)
		}
	default:
		return &root{
			message: message,
			wrapped: e,
			stack:   stack,
		}
	}

	return &wrapped{
		message: message,
		err:     err,
		frame:   frame,
	}
}

// Unwrap returns the result of calling Unwrap() on err if available.
// Matches the behavior of errors.Unwrap in the standard library.
//
// Parameters:
//   - err error: the error to unwrap.
//
// Returns:
//   - error: The next error in the chain, or nil if none.
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}

	return u.Unwrap()
}

// Is reports whether err or any error in its chain matches target.
// Implements an enhanced version of errors.Is from the standard library.
//
// Parameters:
//   - err error: the error to inspect.
//   - target error: the error to compare against.
//
// Returns:
//   - bool: true if any error in err's chain matches target.
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflect.TypeOf(target).Comparable()

	for {
		if isComparable && err == target {
			return true
		}

		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}

		if x, ok := err.(interface{ Is(error) bool }); ok {
			if x.Is(target) {
				return true
			}
		}

		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// As searches err's chain for an error assignable to target and sets target if found.
// Implements an enhanced version of errors.As from the standard library.
//
// Parameters:
//   - err error: the error to inspect.
//   - target interface{}: pointer to the destination interface or concrete type.
//
// Returns:
//   - bool: true if a matching error was found and target was set.
func As(err error, target interface{}) bool {
	if target == nil || err == nil {
		return false
	}

	val := reflect.ValueOf(target)
	typ := val.Type()

	if typ.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	targetType := typ.Elem()

	if targetType.Kind() != reflect.Interface && !targetType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}

	for {
		// Try custom As method first
		if x, ok := err.(interface{ As(interface{}) bool }); ok {
			if x.As(target) {
				return true
			}
		}

		// Standard type assignment check
		if reflect.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(err))

			return true
		}

		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// Cause returns the underlying root cause of the error by recursively unwrapping.
// Unlike Unwrap, it follows the entire chain to the original error.
//
// Parameters:
//   - err error: the error to inspect.
//
// Returns:
//   - error: The deepest non-wrapped error in the chain.
func Cause(err error) error {
	for {
		uerr := Unwrap(err)
		if uerr == nil {
			return err
		}

		err = uerr
	}
}
