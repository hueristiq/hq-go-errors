package errors

// StringFormat holds formatting rules and separators for rendering errors
// as human-readable strings.
//
// Fields:
//   - Options (FormatOptions): Format options (e.g. omitting stack trace or inverting the output order).
//   - MsgStackSep (string): Separator between error messages and stack frame data.
//   - PreStackSep (string): Separator at the beginning of each stack frame.
//   - StackElemSep (string): Separator between elements of each stack frame.
//   - ErrorSep (string): Separator between each error in the chain.
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
func NewDefaultStringFormat(options FormatOptions) (format StringFormat) {
	format = StringFormat{
		Options: options,
	}

	if options.WithTrace {
		format.MsgStackSep = "\n"
		format.PreStackSep = "\t"
		format.StackElemSep = ":"
		format.ErrorSep = "\n"
	} else {
		format.ErrorSep = ": "
	}

	return
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
